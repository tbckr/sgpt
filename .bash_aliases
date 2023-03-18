# These are some bash aliases and functions to make you more productive in your daily work

# short for git summarize
gsum() {
    commit_message="$(sgpt txt "Generate git commit message, my changes: $(git diff)")"
    printf "%s\n" "$commit_message"
    read -rp "Do you want to commit your changes with this commit message? [y/N] " response
    if [[ $response =~ ^[Yy]$ ]]; then
        git add . && git commit -m "$commit_message"
    else
        echo "Commit cancelled."
    fi
}
