import boto3
import os
import json
from datetime import datetime

def datetime_handler(obj):
    if isinstance(obj, datetime):
        return obj.isoformat()
    return None

def list_tables_and_print_details():
    # Create a DynamoDB client
    dynamodb = boto3.client('dynamodb', region_name='eu-north-1')

    # List all tables
    response = dynamodb.list_tables()
    table_names = response.get('TableNames', [])

    # Use absolute path for the file
    file_path = os.path.join('/root/good-blast-real', 'go_tables.txt')

    # Print details for each table and write to file
    try:
        with open(file_path, 'w') as f:
            for table_name in table_names:
                print("\n")
                table_details = dynamodb.describe_table(TableName=table_name)
                print(f"Table Name: {table_name}")
                f.write(json.dumps(table_details, indent=2, default=datetime_handler))
                print(table_details)
                print("\n")
                # Write table name and details to file and flush
                f.write(f"Table Name: {table_name}\n")
                f.write("Details:\n")
                f.write(json.dumps(table_details, indent=2, default=datetime_handler))
                f.write("\n\n")
                f.flush()
    except IOError as e:
        print(f"Error writing to file: {e}")

if __name__ == "__main__":
    list_tables_and_print_details()