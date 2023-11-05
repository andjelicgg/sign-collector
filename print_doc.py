import os
import sys
import subprocess
import argparse

def print_document(file_path):
    # Check if the file exists
    if not os.path.isfile(file_path):
        print("The file does not exist.")
        return

    # Determine the OS and execute the respective print command
    if sys.platform == "darwin":
        # macOS
        try:
            subprocess.run(["lpr", file_path], check=True)
            print("Document sent to the printer on macOS.")
        except subprocess.CalledProcessError as e:
            print(f"An error occurred while trying to print the document on macOS: {e}")
    elif sys.platform == "win32":
        # Windows
        try:
            os.startfile(file_path, 'print')
            print("Document sent to the printer on Windows.")
        except Exception as e:
            print(f"An error occurred while trying to print the document on Windows: {e}")
    else:
        print("Unsupported operating system.")

def main():
    parser = argparse.ArgumentParser(description="Print a document.")
    parser.add_argument("file_path", type=str, help="Path to the document file to print.")
    args = parser.parse_args()

    print_document(args.file_path)

if __name__ == "__main__":
    main()
