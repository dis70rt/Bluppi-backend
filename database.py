import os
import psycopg2

from dotenv import load_dotenv
load_dotenv(override=True)

class SynqItDB:
    def __init__(self):
        self.connection = None
        self.cursor = None

    def __enter__(self):
        self.connection = psycopg2.connect(
            dbname=os.getenv("DB_NAME"),
            user=os.getenv("DB_USER"),
            password=os.getenv("DB_PASSWORD"),
            host=os.getenv("DB_HOST"),
            port=os.getenv("DB_PORT"),
        )
        self.cursor = self.connection.cursor()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if self.cursor:
            self.cursor.close()
        if self.connection:
            self.connection.close()

    def close(self):
        if self.cursor:
            self.cursor.close()
        if self.connection:
            self.connection.close()