#!/bin/bash
# this will install or update sgpt and create a keybinding to cntrl-l

# Usage Overview
# bash sgpt-ubuntu-debian-setup.sh                   # install/update if needed
# bash sgpt-ubuntu-debian-setup.sh --force-install   # always install latest
# bash sgpt-ubuntu-debian-setup.sh --uninstall       # cleanly remove all
# bash sgpt-ubuntu-debian-setup.sh --help            # show help message

# === Configuration ===
TMP_INSTALL_DIR="/tmp/sgpt_install"
TARGET_FILE="/etc/profile.d/sgpt_bind.sh"
# the credential file which will be created (all users can read that file - change permissions as needed or create an sgpt user group)
API_KEY_FILE="/etc/sgpt/openai_key.sh"
# the file with the openai_api_key to be used at installation (root access only) - this file should contain only the key at the first line
SOURCE_CREDENTIAL_FILE="/etc/credentials/sgpt/openai_key"
LOG_FILE="/var/log/sgpt-ubuntu-debian-setup.log"

SGPT_PROFILE_BLOCK="# *** sgpt settings begin ***"
SGPT_PROFILE_CODE='
# *** sgpt settings begin ***
if [ -f /etc/bash.bashrc ]; then
    . /etc/bash.bashrc
fi
# *** sgpt settings end ***
'

SGPT_BASHRC_BLOCK="# *** sgpt settings begin ***"
SGPT_BASHRC_CODE='
# *** sgpt settings begin ***
if [[ $- == *i* ]] && [[ -f /etc/profile.d/sgpt_bind.sh ]]; then
    source /etc/profile.d/sgpt_bind.sh
fi
# *** sgpt settings end ***
'

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
    echo "Usage: ./sgpt-ubuntu-debian-setup.sh [option]"
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

# === Escape helper for sed ===
escape_sed_pattern() {
    echo "$1" | sed -e 's/[][\\/.*^$(){}?+|]/\\&/g'
}

clean_profile() {
ESCAPED_PROFILE=$(escape_sed_pattern "$SGPT_PROFILE_BLOCK")
sudo sed -i "/$ESCAPED_PROFILE/,/# \*\*\* sgpt settings end \*\*\*/d" /etc/profile
}

clean_bashrc() {
ESCAPED_BASHRC=$(escape_sed_pattern "$SGPT_BASHRC_BLOCK")
sudo sed -i "/$ESCAPED_BASHRC/,/# \*\*\* sgpt settings end \*\*\*/d" /etc/bash.bashrc
}


# === UNINSTALL Mode ===
if [[ "$UNINSTALL" == true ]]; then
    echo "🧽 Uninstalling sgpt and cleaning up configuration..."
    sudo rm -f "$TARGET_FILE"
    sudo rm -f "$API_KEY_FILE"
    sudo dpkg -r sgpt || echo "ℹ️ sgpt was not installed."

    clean_profile
    clean_bashrc

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
echo "🔗 Creating binding script $TARGET_FILE ..."
sudo tee "$TARGET_FILE" > /dev/null << EOF
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

sudo chmod +x "$TARGET_FILE"
echo "✅ sgpt_bind.sh created."

# === Step 6: Patch /etc/profile ===
echo "🧩 Updating /etc/profile ..."

clean_profile
echo "$SGPT_PROFILE_CODE" | sudo tee -a /etc/profile > /dev/null

# === Step 7: Patch /etc/bash.bashrc ===
echo "🧩 Updating /etc/bash.bashrc ..."
clean_bashrc
echo "$SGPT_BASHRC_CODE" | sudo tee -a /etc/bash.bashrc > /dev/null

echo "✅ Shell-GPT setup completed successfully."
echo "👉 Open a new terminal or re-login to activate the configuration."
