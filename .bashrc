# These are some bash aliases and functions to make you more productive in your daily work.

## git summarize ##
# Leverage SGPT to produce intelligent and context-sensitive git commit messages.
# By providing one argument, you can define the type of semantic commit (e.g. feat, fix, chore).
# When supplying two arguments, the second parameter allows you to include more details for a more explicit prompt.
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

# Create a alias for access to the GPT-4 Vision API
alias vision='sgpt -m "gpt-4-vision-preview"'

# Create a alias for access to OpenAI's o1-preview model
alias sgpt-o1="sgpt -m \"o1-preview\" --stream=false"
