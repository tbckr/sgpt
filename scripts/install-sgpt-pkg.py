import os
import subprocess
import sys
import datetime

LOG_FILE = "/var/log/install-sgpt-pkg.log"
TMP_INSTALL_DIR = "/tmp/sgpt_install"
PROFILE_SCRIPT_LOCATION = "/etc/profile.d/sgpt_bind.sh"
WRAPPER_SCRIPT_LOCATION = "/usr/bin/sgpt.sh"
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
    run_command(f"rm -f {WRAPPER_SCRIPT_LOCATION}")
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


def create_bind_script():
    print_and_log(f"🔗 Creating binding script {PROFILE_SCRIPT_LOCATION} ...")
    with open(PROFILE_SCRIPT_LOCATION, "w") as f:
        f.write(r"""#!/bin/bash
# Pfad zu sgpt fest verdrahten (Alias ist in Funktionen unzuverlässig)
_sgpt_cmd="/usr/bin/sgpt.sh"
# Nur in interaktiven Bash-Sessions ausführen
if [[ $- == *i* ]] && [[ -n "$BASH_VERSION" ]]; then
  # optional: sicherstellen, dass Readline-Keymap aktiv ist
  set -o emacs 2>/dev/null

  _sgpt_bash() {
    # Wird nur von bind -x aufgerufen; READLINE_LINE ist dann gesetzt
    if [[ -n "$READLINE_LINE" ]]; then
      # Übergibt die aktuelle Zeile an sgpt und ersetzt sie durch die Antwort
      READLINE_LINE="$("$_sgpt_cmd" sh "$READLINE_LINE" --stream)"
      READLINE_POINT=${#READLINE_LINE}
    fi
  }

  # Achtung: überschreibt das übliche Ctrl+L (clear-screen)
  bind -x '"\C-l": _sgpt_bash'
fi

# Alias kann bleiben für normale Nutzung im Terminal
if [[ -x $_sgpt_cmd ]]; then
    alias sgpt='/usr/bin/sgpt.sh'
fi
""")
    os.chmod(PROFILE_SCRIPT_LOCATION, 0o755)
    print_and_log(f"✅ {PROFILE_SCRIPT_LOCATION} created.")


def create_wrapper_script():
    print_and_log(f"🔗 Creating wrapper script {WRAPPER_SCRIPT_LOCATION} ...")
    with open(WRAPPER_SCRIPT_LOCATION, "w") as f:
        f.write("""#!/bin/bash

# Define the path to the configuration file
CONFIG_FILE="$HOME/.config/sgpt/multiconfig.json"

# If the config file doesn't exist, create it with default preset settings
if [[ ! -f "$CONFIG_FILE" ]]; then
  # Create the directory if it doesn't already exist
  mkdir -p "$(dirname "$CONFIG_FILE")"
  
  # Write a default JSON configuration with multiple presets
  cat > "$CONFIG_FILE" <<EOF
{
  "default": {
    "api_base": "https://api.openai.com/v1",
    "api_key": "",
    "model": "gpt-4.1-nano",
    "flags": ""
  },
  "mini": {
    "api_base": "https://api.openai.com/v1",
    "api_key": "",
    "model": "gpt-4.1-mini",
    "flags": ""
  },
  "minilong": {
    "api_base": "https://api.openai.com/v1",
    "api_key": "",
    "model": "gpt-4.1-mini",
    "flags": "-s 100000"
  },
  "claude": {
    "api_base": "https://openrouter.ai/api/v1",
    "api_key": "",
    "model": "anthropic/claude-3.7-sonnet",
    "flags": ""
  }
}
EOF

  # Set secure permissions for the config file (read/write for user only)
  chmod 0600 "$CONFIG_FILE"

  # Notify the user to fill in API keys and flags
  echo "Configuration file created at $CONFIG_FILE. Please set your API keys and flags before proceeding."
  echo "Add as many presets as You want."
  exit 1  # Exit after creating the file
fi

# Check if the user provided a preset with the -p flag
if [[ "$1" == "-p" ]]; then
  PRESET="$2"      # Set the preset from the argument
  shift 2          # Remove -p and preset from the argument list
else
  PRESET="default"  # Use "default" preset if none is specified
fi

# Extract preset settings from the config using `jq`
API_BASE=$(jq -r ".$PRESET.api_base" "$CONFIG_FILE")
API_KEY=$(jq -r ".$PRESET.api_key" "$CONFIG_FILE")
MODEL=$(jq -r ".$PRESET.model" "$CONFIG_FILE")
FLAGS=$(jq -r ".$PRESET.flags" "$CONFIG_FILE")

# Validate that required values are not missing or null
if [[ -z "$API_BASE" || "$API_BASE" == "null" ]]; then
  echo "Error: 'api_base' for preset '$PRESET' is missing."
  echo "Please set it in $CONFIG_FILE"
  exit 1
fi

if [[ -z "$API_KEY" || "$API_KEY" == "null" ]]; then
  echo "Error: 'api_key' for preset '$PRESET' is missing."
  echo "Please set it in $CONFIG_FILE"
  exit 1
fi

if [[ -z "$MODEL" || "$MODEL" == "null" ]]; then
  echo "Error: 'model' for preset '$PRESET' is missing."
  echo "Please set it in $CONFIG_FILE"
  exit 1
fi

# Run the sgpt command with the selected preset's configuration
OPENAI_API_KEY="$API_KEY" OPENAI_API_BASE="$API_BASE" /usr/bin/sgpt -m "$MODEL" $FLAGS "$@"
""")
    os.chmod(WRAPPER_SCRIPT_LOCATION, 0o755)
    print_and_log(f"✅ {WRAPPER_SCRIPT_LOCATION} created.")


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

    create_bind_script()
    create_wrapper_script()

    print_and_log("🧩 Updating /etc/profile ...")
    patch_file("/etc/profile", SGPT_PROFILE_BLOCK_START, SGPT_PROFILE_BLOCK_END, SGPT_PROFILE_CODE)

    print_and_log("🧩 Updating /etc/bash.bashrc ...")
    patch_file("/etc/bash.bashrc", SGPT_BASHRC_BLOCK_START, SGPT_BASHRC_BLOCK_END, SGPT_BASHRC_CODE)

    print_and_log("✅ Shell-GPT setup completed successfully.")
    print_and_log("👉 Open a new terminal or re-login to activate the configuration.")


main()
