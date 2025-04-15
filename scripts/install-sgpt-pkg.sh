#!/bin/bash
# this will install or update sgpt and create a keybinding to cntrl-l

# Usage Overview
# bash install-sgpt-pkg.sh                   # install/update if needed
# bash install-sgpt-pkg.sh --force-install   # always install latest
# bash install-sgpt-pkg.sh --uninstall       # cleanly remove all
# bash install-sgpt-pkg.sh --help            # show help message

# === Configuration ===
TMP_INSTALL_DIR="/tmp/sgpt_install"
PROFILE_SCRIPT_LOCATION="/etc/profile.d/sgpt_bind.sh"
# the credential file which will be created (all users can read that file - change permissions as needed or create an sgpt user group)
API_KEY_FILE="/etc/sgpt/openai_key.sh"
# the file with the openai_api_key to be used at installation (root access only) - this file should contain only the key at the first line
SOURCE_CREDENTIAL_FILE="/etc/credentials/sgpt/openai_key"
LOG_FILE="/var/log/install-sgpt-pkg.log"

SGPT_PROFILE_BLOCK_START="# *** sgpt settings begin ***"
SGPT_PROFILE_BLOCK_END="# *** sgpt settings end ***"
SGPT_PROFILE_CODE='if [ -f /etc/bash.bashrc ]; then
    . /etc/bash.bashrc
fi'

SGPT_BASHRC_BLOCK_START="# *** sgpt settings begin ***"
SGPT_BASHRC_BLOCK_END="# *** sgpt settings end ***"
SGPT_BASHRC_CODE='if [[ $- == *i* ]] && [[ -f /etc/profile.d/sgpt_bind.sh ]]; then
    source /etc/profile.d/sgpt_bind.sh
fi'

# === Sudo Check ===
if [[ "$EUID" -ne 0 ]]; then
    echo "❌ Error: This script must be run as root (with sudo)."
    echo "ℹ️  Please run it again using: sudo $0"
    exit 1
fi

# === Start Logging ===
exec > >(tee -a "$LOG_FILE") 2>&1
echo "🗒️  Logging to $LOG_FILE"
echo "🕒 $(date) - sgpt setup started"

# === Parameter Parsing ===
FORCE_INSTALL=false
UNINSTALL=false

show_help() {
    echo ""
    echo "Shell-GPT System Installer"
    echo "--------------------------"
    echo "Usage: ./install-sgpt-pkg.sh [option]"
    echo ""
    echo "Options:"
    echo "  --force-install     Always install the latest version"
    echo "  --uninstall         Remove sgpt and all configurations"
    echo "  --help              Show this help message"
    echo ""
    echo "Logfile: $LOG_FILE"
    echo ""
    exit 0
}

if [[ "$1" == "--help" ]]; then
    show_help
elif [[ "$1" == "--force-install" ]]; then
    FORCE_INSTALL=true
    echo "⚠️  --force-install enabled: sgpt will be reinstalled."
elif [[ "$1" == "--uninstall" ]]; then
    UNINSTALL=true
    echo "🧹 --uninstall enabled: sgpt and configs will be removed."
fi


# === Remove a block of text from a file ===
remove_block_between_lines() {
    # Arguments:
    #   $1 - The exact starting line (must match fully)
    #   $2 - The exact ending line (must match fully)
    #   $3 - The file path to operate on
    # Returns:
    #   0 - Success
    #   1 - File does not exist
    #   2 - Start line not found
    #   3 - End line not found
    local start="$1"
    local end="$2"
    local file="$3"

    # Check if file exists
    if ! sudo test -f "$file"; then
        echo "Error: File '$file' does not exist."
        return 1
    fi

    # Check if start line exists
    if ! sudo grep -Fxq "$start" "$file"; then
        echo "Error: Start line not found in '$file'."
        return 2
    fi

    # Check if end line exists
    if ! sudo grep -Fxq "$end" "$file"; then
        echo "Error: End line not found in '$file'."
        return 3
    fi

    # Process and replace the file content
    sudo awk -v start="$start" -v end="$end" '
        $0 == start {in_block=1; next}
        $0 == end && in_block {in_block=0; next}
        !in_block
    ' "$file" | sudo tee "$file.tmp" > /dev/null && sudo mv "$file.tmp" "$file"

    return 0
}

remove_sgpt_profile_settings() {
  remove_block_between_lines "${SGPT_PROFILE_BLOCK_START}" "${SGPT_PROFILE_BLOCK_END}" "/etc/profile"
}

remove_sgpt_bashrc_settings() {
remove_block_between_lines "${SGPT_BASHRC_BLOCK_START}" "${SGPT_BASHRC_BLOCK_END}" "/etc/bash.bashrc"
}


# === UNINSTALL Mode ===
if [[ "$UNINSTALL" == true ]]; then
    echo "🧽 Uninstalling sgpt and cleaning up configuration..."
    sudo rm -f "$PROFILE_SCRIPT_LOCATION"
    sudo rm -f "$API_KEY_FILE"
    sudo dpkg -r sgpt || echo "ℹ️ sgpt was not installed."
    remove_sgpt_profile_settings
    remove_sgpt_bashrc_settings
    echo "✅ sgpt successfully uninstalled and cleaned up."
    exit 0
fi

# === Step 1: Check installed version ===
if dpkg -s sgpt &>/dev/null; then
    CURRENT_VERSION=$(dpkg -s sgpt | grep '^Version' | awk '{print $2}')
else
    CURRENT_VERSION="none"
fi
echo "📦 Installed version: $CURRENT_VERSION"

# === Step 2: Get latest version from GitHub ===
LATEST_VERSION=$(curl -s https://api.github.com/repos/tbckr/sgpt/releases/latest | grep '"tag_name":' | cut -d '"' -f 4 | sed 's/^v//')
echo "🌐 Latest available version: $LATEST_VERSION"

# === Step 3: Install or update sgpt ===
if [[ "$CURRENT_VERSION" != "$LATEST_VERSION" || "$FORCE_INSTALL" == true ]]; then
    echo "⬇️  Installing/updating sgpt to version $LATEST_VERSION ..."
    sudo mkdir -p "$TMP_INSTALL_DIR"
    curl -s https://api.github.com/repos/tbckr/sgpt/releases/latest \
    | grep "browser_download_url" \
    | grep ".deb" \
    | grep "amd64" \
    | cut -d '"' -f 4 \
    | xargs -n 1 sudo curl -L -o "$TMP_INSTALL_DIR/sgpt-latest.deb"

    sudo dpkg -i "$TMP_INSTALL_DIR/sgpt-latest.deb"
    sudo rm -rf "$TMP_INSTALL_DIR"
    echo "✅ sgpt installed/updated."
else
    echo "✅ sgpt is already up to date. No installation needed."
fi

# === Step 4: Read API key ===
echo "🔐 Reading API key from $SOURCE_CREDENTIAL_FILE ..."
if sudo test -f "$SOURCE_CREDENTIAL_FILE"; then
    OPENAI_API_KEY=$(sudo cat "$SOURCE_CREDENTIAL_FILE")
    sudo mkdir -p /etc/sgpt
    sudo tee "$API_KEY_FILE" > /dev/null <<EOF
export OPENAI_API_KEY="${OPENAI_API_KEY}"
EOF
    sudo chmod 644 "$API_KEY_FILE"
    sudo chown root:root "$API_KEY_FILE"
    echo "✅ API key saved to $API_KEY_FILE"
else
    echo "❌ Error: API key file not found or not readable. Aborting."
    exit 1
fi

# === Step 5: Create sgpt_bind.sh ===
echo "🔗 Creating binding script $PROFILE_SCRIPT_LOCATION ..."
sudo tee "$PROFILE_SCRIPT_LOCATION" > /dev/null << EOF
# Shell-GPT API key loader
if [[ -f "$API_KEY_FILE" ]]; then
    source "$API_KEY_FILE"
else
    echo "⚠️  Warning: API key file $API_KEY_FILE not found."
fi

# Shell-GPT Bash integration v0.2
_sgpt_bash() {
    if [[ -n "\$READLINE_LINE" ]]; then
        READLINE_LINE=\$(sgpt sh "\$READLINE_LINE" --stream)
        READLINE_POINT=\${#READLINE_LINE}
    fi
}

# Bind Ctrl+L to sgpt
bind -x '"\\C-l": _sgpt_bash'
EOF

sudo chmod +x "$PROFILE_SCRIPT_LOCATION"
echo "✅ sgpt_bind.sh created."

# === Step 6: Patch /etc/profile ===
echo "🧩 Updating /etc/profile ..."

remove_sgpt_profile_settings
echo "$SGPT_PROFILE_BLOCK_START" | sudo tee -a /etc/profile > /dev/null
echo "$SGPT_PROFILE_CODE" | sudo tee -a /etc/profile > /dev/null
echo "$SGPT_PROFILE_BLOCK_END" | sudo tee -a /etc/profile > /dev/null

# === Step 7: Patch /etc/bash.bashrc ===
echo "🧩 Updating /etc/bash.bashrc ..."
remove_sgpt_bashrc_settings

echo "$SGPT_BASHRC_BLOCK_START" | sudo tee -a /etc/bash.bashrc > /dev/null
echo "$SGPT_BASHRC_CODE" | sudo tee -a /etc/bash.bashrc > /dev/null
echo "$SGPT_BASHRC_BLOCK_END" | sudo tee -a /etc/bash.bashrc > /dev/null

echo "✅ Shell-GPT setup completed successfully."
echo "👉 Open a new terminal or re-login to activate the configuration."
