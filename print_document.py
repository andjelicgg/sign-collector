import os
import sys
import subprocess

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
            win32api.ShellExecute(
                0,
                "print",
                file_path,
                '/d:"%s"' % win32print.GetDefaultPrinter(),
                ".",
                0
            )
            print("Document sent to the printer on Windows.")
        except Exception as e:
            print(f"An error occurred while trying to print the document on Windows: {e}")
    else:
        print("Unsupported operating system.")

# Replace '/path/to/document' with the actual file path of your document
print_document('templates/form-01.pdf')
