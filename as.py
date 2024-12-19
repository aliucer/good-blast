import os
import boto3


with open('go_files.txt', 'w') as output_file:
    for root, _, files in os.walk('.'):
        for file in files:
            if file.endswith('.go'):
                relative_path = os.path.join(root, file)
                output_file.write(f"{relative_path}\n")
                with open(relative_path, 'r') as input_file:
                    code = input_file.read()
                    output_file.write(f"{code}\n")