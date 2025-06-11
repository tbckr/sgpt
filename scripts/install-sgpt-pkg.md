# 📄 Documentation: `install-sgpt-pkg.py`

## 📌 Overview

This Python script automates the installation, update, and removal of the **Shell-GPT (sgpt)** tool. It also sets up:

- **Bash integration** (e.g., binding Ctrl+L to call sgpt),
- A **shell wrapper script** for enhanced invocation,
- And a **multi-provider configuration system** per user.

After installation, `sgpt` can be used with different **presets** defined in a user-specific configuration file. Each preset can include:

- Custom **API endpoint** (e.g., OpenAI, OpenRouter),
- **Model name** (e.g., `gpt-4.1-mini`, `claude-3.7-sonnet`),
- User-specific **API key**,
- Additional **flags** (e.g., for streaming, long context).

```bash
# new flag -p to select a preset which can preset all the settings mentioned above 
$> sgpt -p mini "what is the mass of the sun"

# use <cntrl-l> hotkey to get back a command which can be executed with just pressing enter : 
$> list all files < 8kB <cntrl-l>
$> find . -type f -size -8k | ls -l
```


This allows every user on the system to define and use their own AI setup — with complete flexibility over providers, models, context lengths, and access keys.

---

## ⚙️ Requirements

- Root privileges (`sudo`)
- Debian-based Linux system (e.g., Debian, Ubuntu)
- Python 3
- Installed tools: `dpkg`, `curl`, `jq`

---

## ▶️ Usage

```bash
# install or update to the latest sgpt
sudo python3 install-sgpt-pkg.py [OPTION]

# install sgpt (from bitranox fork)
apt-get install jq
curl -sL https://raw.githubusercontent.com/bitranox/sgpt/main/scripts/install-sgpt-pkg.py | python3 -

```

### Options:

| Option            | Description                                                                          |
|-------------------|--------------------------------------------------------------------------------------|
| `--help`          | Show usage instructions                                                              |
| `--force-install` | Force installation of the latest version                                             |
| `--uninstall`     | Uninstall `sgpt` and remove all integration files (keeps user preset configurations) |

---

## 📁 File Paths

| Component             | Path                                    |
|-----------------------|-----------------------------------------|
| Log file              | `/var/log/install-sgpt-pkg.log`         |
| Temporary install dir | `/tmp/sgpt_install`                     |
| Bash bind script      | `/etc/profile.d/sgpt_bind.sh`           |
| Shell wrapper script  | `/usr/bin/sgpt.sh`                      |
| sgpt config file      | `$HOME/.config/sgpt/multiconfig.json`   |

---

## 🛡️ Security Notes

- File permissions are set strictly (`chmod 0600` for config, `0755` for scripts)
- No default API keys included
- User is prompted to edit the config manually
- config survives updates and new installations
- environment is not bloated with keys
- every user on the system can have a different configuration

---

## ✅ Example Usage

```bash
# Fresh installation or update
sudo python3 install-sgpt-pkg.py

# Force reinstallation
sudo python3 install-sgpt-pkg.py --force-install

# Full removal
sudo python3 install-sgpt-pkg.py --uninstall
```

---

## 📌 Notes

- After installation: open a new terminal or re-login  
- `Ctrl+L` is now bound to Shell-GPT  
- Config file `~/.config/sgpt/multiconfig.json` must be edited by the user

```

---
