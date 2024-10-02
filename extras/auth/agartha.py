# -*- coding: utf-8 -*-

"""
Agartha authentication.

Provide authentication using Postgres.
Use Returner database connection settings.

Enable Agartha authentication.

.. code-block:: yaml

    external_auth:
      Agartha:
        paul:
          - test.*

:depends:   - psycopg2.
"""

from __future__ import absolute_import
import logging
import itertools
import sys
from contextlib import contextmanager

from jinja2 import Environment

import salt.utils.data
import salt.utils.stringutils

from salt.exceptions import CommandExecutionError, SaltInvocationError, SaltMasterError


log = logging.getLogger(__name__)

try:
    # pylint: disable=no-name-in-module
    import ldap
    import ldap.filter
    import ldap.modlist

    HAS_LDAP = True
    # pylint: enable=no-name-in-module
except ImportError:
    HAS_LDAP = False

__virtualname__ = "agartha"


def __virtual__():
    if HAS_MYSQL or HAS_POSTGRES:
        return __virtualname__
    return False, "Missing MySQL or Postgres modules"


# Defaults, override in master config
__defopts__ = {
    "auth.ldap.basedn": "",
    "auth.ldap.uri": "",
    "auth.ldap.server": "localhost",
    "auth.ldap.port": "389",
    "auth.ldap.starttls": False,
    "auth.ldap.tls": False,
    "auth.ldap.no_verify": False,
    "auth.ldap.anonymous": False,
    "auth.ldap.scope": 2,
    "auth.ldap.groupou": "Groups",
    "auth.ldap.accountattributename": "memberUid",
    "auth.ldap.groupattribute": "memberOf",
    "auth.ldap.persontype": "person",
    "auth.ldap.groupclass": "posixGroup",
    "auth.ldap.activedirectory": False,
    "auth.ldap.freeipa": False,
    "auth.ldap.minion_stripdomains": [],
}

try:
    import MySQLdb

    HAS_MYSQL = True
except ImportError:
    HAS_MYSQL = False

try:
    import psycopg2

    HAS_POSTGRES = True
except ImportError:
    HAS_POSTGRES = False


def _get_options():
    """
    Returns options used for the database connection.
    """

    _options = {}
    defaults = {
        "host": "salt",
        "user": "salt",
        "pass": "salt",
        "db": "salt",
        "port": 3306,
        "ssl_ca": None,
        "ssl_cert": None,
        "ssl_key": None,
    }

    if HAS_MYSQL:
        for k, v in defaults.items():
            try:
                _options[k] = __opts__["{}.{}".format("mysql", k)]
            except KeyError:
                _options[k] = v
    else:
        # Use "returner.pgupsert" options.
        defaults.pop("pass")
        defaults["passwd"] = "salt"
        defaults["port"] = 5432
        for k, v in defaults.items():
            try:
                _options[k] = __opts__["{}.{}".format("returner.pgupsert", k)]
            except KeyError:
                _options[k] = v

    # post processing
    for k, v in defaults.items():
        if isinstance(v, str) and v.lower() == "none":
            # Ensure 'None' is rendered as None
            _options[k] = None
        if k == "port":
            # Ensure port is an int
            _options[k] = int(v)

    return _options


def _config(key, mandatory=True, opts=None):
    """
    Return a value for 'name' from master config file options or defaults.
    """
    try:
        if opts:
            value = opts["auth.ldap.{}".format(key)]
        else:
            value = __opts__["auth.ldap.{}".format(key)]
    except KeyError:
        try:
            value = __defopts__["auth.ldap.{}".format(key)]
        except KeyError:
            if mandatory:
                msg = "missing auth.ldap.{} in master config".format(key)
                raise SaltInvocationError(msg)
            return False
    return value


def _render_template(param, username):
    """
    Render config template, substituting username where found.
    """
    env = Environment()
    template = env.from_string(param)
    variables = {"username": username}
    return template.render(variables)


class _LDAPConnection:
    """
    Setup an LDAP connection.
    """

    def __init__(
        self,
        uri,
        server,
        port,
        starttls,
        tls,
        no_verify,
        binddn,
        bindpw,
        anonymous,
        accountattributename,
        activedirectory=False,
    ):
        """
        Bind to an LDAP directory using passed credentials.
        """
        self.uri = uri
        self.server = server
        self.port = port
        self.starttls = starttls
        self.tls = tls
        self.binddn = binddn
        self.bindpw = bindpw
        if not HAS_LDAP:
            raise CommandExecutionError(
                "LDAP connection could not be made, the python-ldap module is "
                "not installed. Install python-ldap to use LDAP external auth."
            )
        if self.starttls and self.tls:
            raise CommandExecutionError(
                "Cannot bind with both starttls and tls enabled."
                "Please enable only one of the protocols"
            )

        schema = "ldaps" if tls else "ldap"
        if self.uri == "":
            self.uri = "{}://{}:{}".format(schema, self.server, self.port)

        try:
            if no_verify:
                ldap.set_option(ldap.OPT_X_TLS_REQUIRE_CERT, ldap.OPT_X_TLS_NEVER)

            self.ldap = ldap.initialize("{}".format(self.uri))
            self.ldap.protocol_version = 3  # ldap.VERSION3
            self.ldap.set_option(ldap.OPT_REFERRALS, 0)  # Needed for AD

            if not anonymous:
                if not self.bindpw:
                    raise CommandExecutionError(
                        "LDAP bind password is not set: password cannot be empty if auth.ldap.anonymous is False"
                    )
                if self.starttls:
                    self.ldap.start_tls_s()
                self.ldap.simple_bind_s(self.binddn, self.bindpw)
        except Exception as ldap_error:  # pylint: disable=broad-except
            raise CommandExecutionError(
                "Failed to bind to LDAP server {} as {}: {}".format(
                    self.uri, self.binddn, ldap_error
                )
            )


def _bind_for_search(anonymous=False, opts=None):
    """
    Bind with binddn and bindpw only for searching LDAP
    :param anonymous: Try binding anonymously
    :param opts: Pass in when __opts__ is not available
    :return: LDAPConnection object
    """
    # Get config params; create connection dictionary
    connargs = {}
    # config params (auth.ldap.*)
    params = {
        "mandatory": [
            "uri",
            "server",
            "port",
            "starttls",
            "tls",
            "no_verify",
            "anonymous",
            "accountattributename",
            "activedirectory",
        ],
        "additional": [
            "binddn",
            "bindpw",
            "filter",
            "groupclass",
            "auth_by_group_membership_only",
        ],
    }

    paramvalues = {}

    for param in params["mandatory"]:
        paramvalues[param] = _config(param, opts=opts)

    for param in params["additional"]:
        paramvalues[param] = _config(param, mandatory=False, opts=opts)

    paramvalues["anonymous"] = anonymous

    # Only add binddn/bindpw to the connargs when they're set, as they're not
    # mandatory for initializing the LDAP object, but if they're provided
    # initially, a bind attempt will be done during the initialization to
    # validate them
    if paramvalues["binddn"]:
        connargs["binddn"] = paramvalues["binddn"]
        if paramvalues["bindpw"]:
            params["mandatory"].append("bindpw")

    for name in params["mandatory"]:
        connargs[name] = paramvalues[name]

    if not paramvalues["anonymous"]:
        if paramvalues["binddn"] and paramvalues["bindpw"]:
            # search for the user's DN to be used for the actual authentication
            return _LDAPConnection(**connargs).ldap


def _bind(username, anonymous=False, opts=None):
    """
    Authenticate via an LDAP bind
    """
    # Get config params; create connection dictionary
    basedn = _config("basedn", opts=opts)
    scope = _config("scope", opts=opts)
    connargs = {}
    # config params (auth.ldap.*)
    params = {
        "mandatory": [
            "uri",
            "server",
            "port",
            "starttls",
            "tls",
            "no_verify",
            "anonymous",
            "accountattributename",
            "activedirectory",
        ],
        "additional": [
            "binddn",
            "bindpw",
            "filter",
            "groupclass",
            "auth_by_group_membership_only",
        ],
    }

    paramvalues = {}

    for param in params["mandatory"]:
        paramvalues[param] = _config(param, opts=opts)

    for param in params["additional"]:
        paramvalues[param] = _config(param, mandatory=False, opts=opts)

    paramvalues["anonymous"] = anonymous
    if paramvalues["binddn"]:
        # the binddn can also be composited, e.g.
        #   - {{ username }}@domain.com
        #   - cn={{ username }},ou=users,dc=company,dc=tld
        # so make sure to render it first before using it
        paramvalues["binddn"] = _render_template(paramvalues["binddn"], username)
        paramvalues["binddn"] = ldap.filter.escape_filter_chars(paramvalues["binddn"])

    if paramvalues["filter"]:
        escaped_username = ldap.filter.escape_filter_chars(username)
        paramvalues["filter"] = _render_template(
            paramvalues["filter"], escaped_username
        )

    # Only add binddn/bindpw to the connargs when they're set, as they're not
    # mandatory for initializing the LDAP object, but if they're provided
    # initially, a bind attempt will be done during the initialization to
    # validate them
    if paramvalues["binddn"]:
        connargs["binddn"] = paramvalues["binddn"]
        if paramvalues["bindpw"]:
            params["mandatory"].append("bindpw")

    for name in params["mandatory"]:
        connargs[name] = paramvalues[name]

    if not paramvalues["anonymous"]:
        if paramvalues["binddn"] and paramvalues["bindpw"]:
            # search for the user's DN to be used for the actual authentication
            _ldap = _LDAPConnection(**connargs).ldap
            log.debug(
                "Running LDAP user dn search with filter:%s, dn:%s, scope:%s",
                paramvalues["filter"],
                basedn,
                scope,
            )
            result = _ldap.search_s(basedn, int(scope), paramvalues["filter"])
            if not result:
                log.warning("Unable to find user %s", username)
                return False
            elif len(result) > 1:
                # Active Directory returns something odd.  Though we do not
                # chase referrals (ldap.set_option(ldap.OPT_REFERRALS, 0) above)
                # it still appears to return several entries for other potential
                # sources for a match.  All these sources have None for the
                # CN (ldap array return items are tuples: (cn, ldap entry))
                # But the actual CNs are at the front of the list.
                # So with some list comprehension magic, extract the first tuple
                # entry from all the results, create a list from those,
                # and count the ones that are not None.  If that total is more than one
                # we need to error out because the ldap filter isn't narrow enough.
                cns = [tup[0] for tup in result]
                total_not_none = sum(1 for c in cns if c is not None)
                if total_not_none > 1:
                    log.error(
                        "LDAP lookup found multiple results for user %s", username
                    )
                    return False
                elif total_not_none == 0:
                    log.error(
                        "LDAP lookup--unable to find CN matching user %s", username
                    )
                    return False
    return True


def _ldap_user(username):
    """
    Simple LDAP auth
    """
    if not HAS_LDAP:
        log.error("LDAP authentication requires python-ldap module")
        return False

    bind = None

    # If bind credentials are configured, verify that we receive a valid bind
    if _config("binddn", mandatory=False) and _config("bindpw", mandatory=False):
        search_bind = _bind_for_search(anonymous=_config("anonymous", mandatory=False))

        # If username & password are not None, attempt to verify they are valid
        if search_bind and username:
            bind = _bind(
                username,
                anonymous=_config("auth_by_group_membership_only", mandatory=False)
                and _config("anonymous", mandatory=False),
            )
    else:
        bind = _bind(
            username,
            anonymous=_config("auth_by_group_membership_only", mandatory=False)
            and _config("anonymous", mandatory=False),
        )

    if bind:
        log.debug("LDAP authentication successful")
        return bind

    log.error("LDAP _bind authentication FAILED")
    return False


@contextmanager
def _get_serv():
    """
    Return a database cursor
    """
    _options = _get_options()

    log.debug("Generating new DB connection pool")
    if HAS_MYSQL:
        try:
            # An empty ssl_options dictionary passed to MySQLdb.connect will
            # effectively connect w/o SSL.
            ssl_options = {}
            if _options.get("ssl_ca"):
                ssl_options["ca"] = _options.get("ssl_ca")
            if _options.get("ssl_cert"):
                ssl_options["cert"] = _options.get("ssl_cert")
            if _options.get("ssl_key"):
                ssl_options["key"] = _options.get("ssl_key")
            conn = MySQLdb.connect(
                host=_options.get("host"),
                user=_options.get("user"),
                passwd=_options.get("pass"),
                db=_options.get("db"),
                port=_options.get("port"),
                ssl=ssl_options,
            )

        except MySQLdb.connections.OperationalError as exc:
            raise SaltMasterError(
                "MySQL returner could not connect to database: {exc}".format(exc=exc)
            )

        cursor = conn.cursor()

        try:
            yield cursor
        except MySQLdb.DatabaseError as err:
            error = err.args
            sys.stderr.write(str(error))
            raise err
    else:
        try:
            conn = psycopg2.connect(
                host=_options.get("host"),
                user=_options.get("user"),
                password=_options.get("passwd"),
                database=_options.get("db"),
                port=_options.get("port"),
            )

        except psycopg2.OperationalError as exc:
            raise SaltMasterError(
                "postgres returner could not connect to database: {exc}".format(exc=exc)
            )

        cursor = conn.cursor()

        try:
            yield cursor
        except psycopg2.DatabaseError as err:
            error = err.args
            sys.stderr.write(str(error))
            raise err
        finally:
            conn.close()


def groups(username, **kwargs):
    """
    Authenticate against an LDAP group

    Behavior is highly dependent on if Active Directory is in use.

    AD handles group membership very differently than OpenLDAP.
    See the :ref:`External Authentication <acl-eauth>` documentation for a thorough
    discussion of available parameters for customizing the search.

    OpenLDAP allows you to search for all groups in the directory
    and returns members of those groups.  Then we check against
    the username entered.

    """
    group_list = []

    # If bind credentials are configured, use them instead of user's
    if _config("binddn", mandatory=False) and _config("bindpw", mandatory=False):
        bind = _bind_for_search(anonymous=_config("anonymous", mandatory=False))
    else:
        bind = _bind(
            username,
            kwargs.get("password", ""),
            anonymous=_config("auth_by_group_membership_only", mandatory=False)
            and _config("anonymous", mandatory=False),
        )

    if bind:
        log.debug("ldap bind to determine group membership succeeded!")

        if _config("activedirectory"):
            try:
                get_user_dn_search = "(&({}={})(objectClass={}))".format(
                    _config("accountattributename"), username, _config("persontype")
                )
                user_dn_results = bind.search_s(
                    _config("basedn"),
                    ldap.SCOPE_SUBTREE,
                    get_user_dn_search,
                    ["distinguishedName"],
                )
            except Exception as e:  # pylint: disable=broad-except
                log.error("Exception thrown while looking up user DN in AD: %s", e)
                return group_list
            if not user_dn_results:
                log.error("Could not get distinguished name for user %s", username)
                return group_list
            # LDAP results are always tuples.  First entry in the tuple is the DN
            dn = ldap.filter.escape_filter_chars(user_dn_results[0][0])
            ldap_search_string = "(&(member={})(objectClass={}))".format(
                dn, _config("groupclass")
            )
            log.debug("Running LDAP group membership search: %s", ldap_search_string)
            try:
                search_results = bind.search_s(
                    _config("basedn"),
                    ldap.SCOPE_SUBTREE,
                    ldap_search_string,
                    [
                        salt.utils.stringutils.to_str(_config("accountattributename")),
                        "cn",
                    ],
                )
            except Exception as e:  # pylint: disable=broad-except
                log.error(
                    "Exception thrown while retrieving group membership in AD: %s", e
                )
                return group_list
            for _, entry in search_results:
                if "cn" in entry:
                    group_list.append(salt.utils.stringutils.to_unicode(entry["cn"][0]))
            log.debug("User %s is a member of groups: %s", username, group_list)

        elif _config("freeipa"):
            escaped_username = ldap.filter.escape_filter_chars(username)
            search_base = _config("group_basedn")
            search_string = _render_template(_config("group_filter"), escaped_username)
            search_results = bind.search_s(
                search_base,
                ldap.SCOPE_SUBTREE,
                search_string,
                [
                    salt.utils.stringutils.to_str(_config("accountattributename")),
                    salt.utils.stringutils.to_str(_config("groupattribute")),
                    "cn",
                ],
            )

            for entry, result in search_results:
                for user in itertools.chain(
                    result.get(_config("accountattributename"), []),
                    result.get(_config("groupattribute"), []),
                ):
                    if (
                        username
                        == salt.utils.stringutils.to_unicode(user)
                        .split(",")[0]
                        .split("=")[-1]
                    ):
                        group_list.append(entry.split(",")[0].split("=")[-1])

            log.debug("User %s is a member of groups: %s", username, group_list)

            if not _ldap_user(username):
                log.error("LDAP username and password do not match")
                return []
        else:
            if _config("groupou"):
                search_base = "ou={},{}".format(_config("groupou"), _config("basedn"))
            else:
                search_base = "{}".format(_config("basedn"))
            search_string = "(&({}={})(objectClass={}))".format(
                _config("accountattributename"), username, _config("groupclass")
            )
            search_results = bind.search_s(
                search_base,
                ldap.SCOPE_SUBTREE,
                search_string,
                [
                    salt.utils.stringutils.to_str(_config("accountattributename")),
                    "cn",
                    salt.utils.stringutils.to_str(_config("groupattribute")),
                ],
            )
            for _, entry in search_results:
                if username in salt.utils.data.decode(
                    entry[_config("accountattributename")]
                ):
                    group_list.append(salt.utils.stringutils.to_unicode(entry["cn"][0]))
            for user, entry in search_results:
                if (
                    username
                    == salt.utils.stringutils.to_unicode(user)
                    .split(",")[0]
                    .split("=")[-1]
                ):
                    for group in salt.utils.data.decode(
                        entry[_config("groupattribute")]
                    ):
                        group_list.append(
                            salt.utils.stringutils.to_unicode(group)
                            .split(",")[0]
                            .split("=")[-1]
                        )
            log.debug("User %s is a member of groups: %s", username, group_list)

            # Only test user auth on first call for job.
            # 'show_jid' only exists on first payload so we can use that for the conditional.
            if "show_jid" in kwargs and not _bind(
                username,
                kwargs.get("password"),
                anonymous=_config("auth_by_group_membership_only", mandatory=False)
                and _config("anonymous", mandatory=False),
            ):
                log.error("LDAP username and password do not match")
                return []
    else:
        log.error("ldap bind to determine group membership FAILED!")

    return group_list


def process_acl(auth_list, opts=None):
    """
    Query LDAP, retrieve list of minion_ids from an OU or other search.
    For each minion_id returned from the LDAP search, copy the perms
    matchers into the auth dictionary
    :param auth_list:
    :param opts: __opts__ for when __opts__ is not injected
    :return: Modified auth list.
    """
    ou_names = []
    for item in auth_list:
        if isinstance(item, str):
            continue
        ou_names.extend(
            [
                potential_ou
                for potential_ou in item.keys()
                if potential_ou.startswith("ldap(")
            ]
        )
    if ou_names:
        auth_list = __expand_ldap_entries(auth_list, opts)
    return auth_list


def __expand_ldap_entries(entries, opts=None):
    """

    :param entries: ldap subtree in external_auth config option
    :param opts: Opts to use when __opts__ not defined
    :return: Dictionary with all allowed operations

    Takes the ldap subtree in the external_auth config option and expands it
    with actual minion names

    webadmins%:  <all users in the AD 'webadmins' group>
      - server1
          - .*
      - ldap(OU=webservers,dc=int,dc=bigcompany,dc=com)
        - test.ping
        - service.restart
      - ldap(OU=Domain Controllers,dc=int,dc=bigcompany,dc=com)
        - allowed_fn_list_attribute^

    This function only gets called if auth.ldap.activedirectory = True
    """
    bind = _bind_for_search(opts=opts)
    acl_tree = []
    for user_or_group_dict in entries:
        if not isinstance(user_or_group_dict, dict):
            acl_tree.append(user_or_group_dict)
            continue
        for minion_or_ou, matchers in user_or_group_dict.items():
            permissions = matchers
            retrieved_minion_ids = []
            if minion_or_ou.startswith("ldap("):
                search_base = minion_or_ou.lstrip("ldap(").rstrip(")")

                search_string = "(objectClass=computer)"
                try:
                    search_results = bind.search_s(
                        search_base, ldap.SCOPE_SUBTREE, search_string, ["cn"]
                    )
                    for ldap_match in search_results:
                        try:
                            minion_id = ldap_match[1]["cn"][0].lower()
                            # Some LDAP/AD trees only have the FQDN of machines
                            # in their computer lists.  auth.minion_stripdomains
                            # lets a user strip off configured domain names
                            # and arrive at the basic minion_id
                            if opts.get("auth.ldap.minion_stripdomains", None):
                                for domain in opts["auth.ldap.minion_stripdomains"]:
                                    if minion_id.endswith(domain):
                                        minion_id = minion_id[: -len(domain)]
                                        break
                            retrieved_minion_ids.append(minion_id)
                        except TypeError:
                            # TypeError here just means that one of the returned
                            # entries didn't match the format we expected
                            # from LDAP.
                            pass

                    for minion_id in retrieved_minion_ids:
                        acl_tree.append({minion_id: permissions})
                    log.trace("Expanded acl_tree is: %s", acl_tree)
                except ldap.NO_SUCH_OBJECT:
                    pass
            else:
                acl_tree.append({minion_or_ou: matchers})

    log.trace("__expand_ldap_entries: %s", acl_tree)
    return acl_tree


def auth(username, password):
    """
    Authenticate using a MySQL/Postgres user table
    """
    log.info("Authenticating %s", username)
    if str(password):
        with _get_serv() as cur:
            sql = """
                SELECT a.is_active, a.is_superuser 
                FROM user_settings c 
                INNER JOIN auth_user a ON c.user_id = a.id 
                WHERE a.username = %s AND a.is_active 
                AND c.token = crypt(%s, c.token);
            """
            cur.execute(sql, (username, password))

            result = cur.fetchone()
            if result:
                log.debug("Agartha authentication successful.")
                is_active, is_superuser = result
                if not is_active:
                    log.debug("User failed is_active check.")
                    return False
                if is_superuser:
                    log.debug(
                        "User is_active and is_superuser. Skipping ldap authorization."
                    )
                    return True
        log.debug("Attempting ldap authorization.")
        return _ldap_user(username)
    return False
