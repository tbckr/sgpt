# These are some bash aliases and functions to make you more productive in your daily work

# gsum aka git summarize
gsum() {
    if [ $# -eq 2 ]; then
        query="Generate git commit message using semantic versioning. Declare commit message as $1. $2. My changes: $(git diff)"
    elif [ $# -eq 1 ]; then
        query="Generate git commit message using semantic versioning. Declare commit message as $1. My changes: $(git diff)"
    else
        query="Generate git commit message using semantic versioning. My changes: $(git diff)"
    fi
    commit_message="$(sgpt txt "$query")"
    printf "%s\n" "$commit_message"
    read -rp "Do you want to commit your changes with this commit message? [y/N] " response
    if [[ $response =~ ^[Yy]$ ]]; then
        git add . && git commit -m "$commit_message"
    else
        echo "Commit cancelled."
    fi
}
