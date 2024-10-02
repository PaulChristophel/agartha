import yaml
import psycopg2
import time


def read_config(yaml_file):
    with open(yaml_file, "r") as file:
        config = yaml.safe_load(file)
    return config


def check_table_exists(config):
    try:
        connection = psycopg2.connect(
            dbname=config["cache.pgjsonb.dbname"],
            user=config["cache.pgjsonb.user"],
            password=config["cache.pgjsonb.password"],
            host=config["cache.pgjsonb.host"],
            port=config["cache.pgjsonb.port"],
        )
        cursor = connection.cursor()
        cursor.execute(
            "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name=%s);",
            ("salt_cache",),
        )
        exists = cursor.fetchone()[0]
        cursor.close()
        connection.close()
        return exists
    except psycopg2.Error as error:
        print(f"Database connection error: {error}")
        return False


def wait_for_table(yaml_file):
    config = read_config(yaml_file)
    while True:
        if check_table_exists(config):
            return True
        print("Table 'salt_cache' does not exist yet. Waiting...")
        time.sleep(5)


if __name__ == "__main__":
    yaml_file = "/etc/salt/master"
    if wait_for_table(yaml_file):
        print("Table 'salt_cache' exists!")
