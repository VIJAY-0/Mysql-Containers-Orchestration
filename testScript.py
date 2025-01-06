import mysql.connector
from time import sleep

# Connection parameters
host = "localhost"  # e.g., "localhost"
user = "root"  # e.g., "root"
password = "rootpassword"  # e.g., "password"
port = 3305  # Custom port

# Function to test connection
def test_connection():
    try:
        conn = mysql.connector.connect(
            host=host,
            user=user,
            password=password,
            port=port
        )
        cursor = conn.cursor()
        # cursor.execute("SHOW VARIABLES LIKE 'server_id' ;")
        cursor.execute("SHOW DATABASES ;")

        for x in cursor.fetchall():
            print(x)
        conn.close()
        print("Connection successful")
    except mysql.connector.Error as err:
        print(f"Connection failed: {err}")

# Test connection 100 times
for i in range(100):
    print(f"Testing connection {i + 1}...")
    test_connection()
    # sleep(0.01)  # Pause for a second between tests to avoid overload
