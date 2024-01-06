# GPT-4 Vision API

SGPT additionally facilitates the utilization of the GPT-4 Vision API. Include input images using the `-i` or `--input`
flag, supporting both URLs and local images.

```shell
$ sgpt -m "gpt-4-vision-preview" -i "https://upload.wikimedia.org/wikipedia/en/c/cb/Marvin_%28HHGG%29.jpg" "what can you see on the picture?"
The image shows a figure resembling a robot with a humanoid form. It has a
$ sgpt -m "gpt-4-vision-preview" -i pkg/fs/testdata/marvin.jpg "what can you see on the picture?"
The image shows a figure resembling a robot with a sleek, metallic surface. It
```

It is also possible to combine URLs and local images:

```shell
$ sgpt-m "gpt-4-vision-preview" -i "https://upload.wikimedia.org/wikipedia/en/c/cb/Marvin_%28HHGG%29.jpg" -i pkg/fs/testdata/marvin.jpg "what is the difference between those two pictures"
The two images provided appear to be identical. Both show the same depiction of a
```

To avoid specifying the `-m "gpt-4-vision-preview"` for each request, you can streamline the process by creating a bash
alias:

```shell
alias vision='sgpt -m "gpt-4-vision-preview"'
```

For more bash examples, see [.bashrc](https://github.com/tbckr/sgpt/blob/main/.bashrc).

**Important:** The GPT-4-vision API integration is currently in beta and may change in the future.
