import os
import subprocess
import sys
import datetime

### TODO
# Here is a workaround to avoid using an environment variable.
#
# Add the directory with the following content into your PATH:
#
# ./
# ├── config.json
# └── sgpt.sh --> put in /usr/local/bin - which is in the path before /usr/bin/sgpt
# $ cat config.json
# {"openai_api_key": "<value>"}
#
# $ cat sgpt.sh
# OPENAI_API_KEY=$(cat $(dirname ${BASH_SOURCE[0]})/config.json | jq -r .openai_api_key) sgpt "$@"
# Now you can use sgpt.sh instead of sgpt to run your scripts/commands without exposing the API key.
#
###


LOG_FILE = "/var/log/install-sgpt-pkg.log"
TMP_INSTALL_DIR = "/tmp/sgpt_install"
PROFILE_SCRIPT_LOCATION = "/etc/profile.d/sgpt_bind.sh"
# the credential file which will be created - all users can read that file
# change permissions as needed or create a sgpt user group
API_KEY_FILE = "/etc/sgpt/openai_key.sh"
# the file with the openai_api_key to be used at installation (root access only)
# this file have to contain only the key as the first line
SOURCE_CREDENTIAL_FILE = "/etc/credentials/sgpt/openai_key"
SGPT_PROFILE_BLOCK_START = "# *** sgpt settings begin ***"
SGPT_PROFILE_BLOCK_END = "# *** sgpt settings end ***"
SGPT_PROFILE_CODE = '''if [ -f /etc/bash.bashrc ]; then
    . /etc/bash.bashrc
fi
'''
SGPT_BASHRC_BLOCK_START = "# *** sgpt settings begin ***"
SGPT_BASHRC_BLOCK_END = "# *** sgpt settings end ***"
SGPT_BASHRC_CODE = '''if [[ $- == *i* ]] && [[ -f /etc/profile.d/sgpt_bind.sh ]]; then
    source /etc/profile.d/sgpt_bind.sh
fi
'''


def is_root():
    return os.geteuid() == 0


def log_start():
    with open(LOG_FILE, "a") as f:
        f.write(f"\n🗒️  Logging to {LOG_FILE}\n")
        f.write(f"🕒 {datetime.datetime.now()} - sgpt setup started\n")


def print_and_log(msg):
    with open(LOG_FILE, "a") as f:
        f.write(msg + "\n")
    print(msg)


def parse_arguments():
    force_install = False
    uninstall = False

    if "--help" in sys.argv:
        show_help()
    if "--force-install" in sys.argv:
        force_install = True
        print_and_log("⚠️  --force-install enabled: sgpt will be reinstalled.")
    if "--uninstall" in sys.argv:
        uninstall = True
        print_and_log("🧹 --uninstall enabled: sgpt and configs will be removed.")
    return force_install, uninstall


def show_help():
    print(f"""
Shell-GPT System Installer
--------------------------
Usage: sudo python3 install-sgpt-pkg.py [option]

Options:
  --force-install     Always install the latest version
  --uninstall         Remove sgpt and all configurations
  --help              Show this help message

Logfile: {LOG_FILE}
""")
    sys.exit(0)


def run_command(cmd, check=True):
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    if check and result.returncode != 0:
        print_and_log(f"❌ Command failed: {cmd}\n{result.stderr}")
        sys.exit(result.returncode)
    return result.stdout.strip()


def get_current_version():
    try:
        return run_command("dpkg -s sgpt | grep '^Version' | awk '{print $2}'")
    except Exception:
        return "none"


def get_latest_version():
    output = run_command(
        "curl -s https://api.github.com/repos/tbckr/sgpt/releases/latest | grep '\"tag_name\"' | cut -d '\"' -f 4 | sed 's/^v//'"
    )
    return output


def install_sgpt():
    print_and_log(f"⬇️  Installing/updating sgpt...")
    os.makedirs(TMP_INSTALL_DIR, exist_ok=True)
    download_url = run_command(
        "curl -s https://api.github.com/repos/tbckr/sgpt/releases/latest | "
        "grep browser_download_url | grep .deb | grep amd64 | cut -d '\"' -f 4"
    )
    run_command(f"curl -L {download_url} -o {TMP_INSTALL_DIR}/sgpt-latest.deb")
    run_command(f"dpkg -i {TMP_INSTALL_DIR}/sgpt-latest.deb")
    run_command(f"rm -rf {TMP_INSTALL_DIR}")
    print_and_log("✅ sgpt installed/updated.")


def uninstall_sgpt():
    print_and_log("🧽 Uninstalling sgpt and cleaning up configuration...")
    run_command(f"rm -f {PROFILE_SCRIPT_LOCATION}")
    run_command(f"rm -f {API_KEY_FILE}")
    run_command("dpkg -r sgpt || true")
    remove_block_between_lines(SGPT_PROFILE_BLOCK_START, SGPT_PROFILE_BLOCK_END, "/etc/profile")
    remove_block_between_lines(SGPT_BASHRC_BLOCK_START, SGPT_BASHRC_BLOCK_END, "/etc/bash.bashrc")
    print_and_log("✅ sgpt successfully uninstalled and cleaned up.")
    sys.exit(0)


def remove_block_between_lines(start, end, file_path):
    if not os.path.exists(file_path):
        return
    with open(file_path, "r") as f:
        lines = f.readlines()
    with open(file_path, "w") as f:
        in_block = False
        for line in lines:
            if start in line:
                in_block = True
                continue
            if end in line and in_block:
                in_block = False
                continue
            if not in_block:
                f.write(line)


def read_api_key():
    print_and_log(f"🔐 Reading API key from {SOURCE_CREDENTIAL_FILE} ...")
    if os.path.exists(SOURCE_CREDENTIAL_FILE):
        with open(SOURCE_CREDENTIAL_FILE, "r") as f:
            key = f.readline().strip()
        os.makedirs("/etc/sgpt", exist_ok=True)
        with open(API_KEY_FILE, "w") as f:
            f.write(f'export OPENAI_API_KEY="{key}"\n')
        os.chmod(API_KEY_FILE, 0o644)
        print_and_log(f"✅ API key saved to {API_KEY_FILE}")
    else:
        print_and_log("❌ Error: API key file not found or not readable. Aborting.")
        sys.exit(1)


def create_bind_script():
    print_and_log(f"🔗 Creating binding script {PROFILE_SCRIPT_LOCATION} ...")
    with open(PROFILE_SCRIPT_LOCATION, "w") as f:
        f.write(f"""# Shell-GPT API key loader
if [[ -f "{API_KEY_FILE}" ]]; then
    source "{API_KEY_FILE}"
else
    echo "⚠️  Warning: API key file {API_KEY_FILE} not found."
fi

# Shell-GPT Bash integration v0.2
_sgpt_bash() {{
    if [[ -n "$READLINE_LINE" ]]; then
        READLINE_LINE=$(sgpt sh "$READLINE_LINE" --stream)
        READLINE_POINT=${{#READLINE_LINE}}
    fi
}}

# Bind Ctrl+L to sgpt
bind -x '"\\C-l": _sgpt_bash'
""")
    os.chmod(PROFILE_SCRIPT_LOCATION, 0o755)
    print_and_log("✅ sgpt_bind.sh created.")


def patch_file(file_path, start_tag, end_tag, content):
    remove_block_between_lines(start_tag, end_tag, file_path)
    with open(file_path, "a") as f:
        f.write(f"{start_tag}\n{content}{end_tag}\n")


def main():
    if not is_root():
        print("❌ Error: This script must be run as root (with sudo).")
        print(f"ℹ️  Please run it again using: sudo {sys.argv[0]}")
        sys.exit(1)

    os.makedirs(os.path.dirname(LOG_FILE), exist_ok=True)
    log_start()

    force_install, uninstall = parse_arguments()

    if uninstall:
        uninstall_sgpt()

    current_version = get_current_version()
    latest_version = get_latest_version()

    print_and_log(f"📦 Installed version: {current_version}")
    print_and_log(f"🌐 Latest available version: {latest_version}")

    if force_install or current_version != latest_version:
        install_sgpt()
    else:
        print_and_log("✅ sgpt is already up to date. No installation needed.")

    read_api_key()
    create_bind_script()

    print_and_log("🧩 Updating /etc/profile ...")
    patch_file("/etc/profile", SGPT_PROFILE_BLOCK_START, SGPT_PROFILE_BLOCK_END, SGPT_PROFILE_CODE)

    print_and_log("🧩 Updating /etc/bash.bashrc ...")
    patch_file("/etc/bash.bashrc", SGPT_BASHRC_BLOCK_START, SGPT_BASHRC_BLOCK_END, SGPT_BASHRC_CODE)

    print_and_log("✅ Shell-GPT setup completed successfully.")
    print_and_log("👉 Open a new terminal or re-login to activate the configuration.")


main()
