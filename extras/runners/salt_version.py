import salt.utils.functools
import salt.utils.versions


def version():
    """
    Return the version of salt on the minion

    CLI Example:

    .. code-block:: bash

        salt '*' test.version
    """
    return salt.version.__version__


def versions_information():
    """
    Report the versions of dependent and system software

    CLI Example:

    .. code-block:: bash

        salt '*' test.versions_information
    """
    return salt.version.versions_information()


def versions_report():
    """
    Returns versions of components used by salt

    CLI Example:

    .. code-block:: bash

        salt '*' test.versions_report
    """
    return "\n".join(salt.version.versions_report())


versions = salt.utils.functools.alias_function(versions_report, "versions")
