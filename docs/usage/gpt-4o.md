# GPT-4o and GPT-4 Vision API Support

SGPT additionally facilitates the utilization of the GPT-4o and GPT-4 Vision API. Include input images using the `-i`
or `--input` flag, supporting both URLs and local images.

```shell
$ sgpt -m "gpt-4o" -i pkg/fs/testdata/marvin.jpg "what can you see on the picture?"
The picture shows a robot with a large, round head and an expressive, downward-slanting triangular eye. The body of the robot is designed with a sleek, somewhat shiny, metallic structure and it is pointing with its right hand. The design appears to be humanoid with distinct arms, legs, and a segmented torso.
$ sgpt -m "gpt-4-vision-preview" -i "https://upload.wikimedia.org/wikipedia/en/c/cb/Marvin_%28HHGG%29.jpg" "what can you see on the picture?"
The image shows a figure resembling a robot with a humanoid form. It has a
```

It is also possible to combine URLs and local images:

```shell
$ sgpt -m "gpt-4o" -i "https://upload.wikimedia.org/wikipedia/en/c/cb/Marvin_%28HHGG%29.jpg" -i pkg/fs/testdata/marvin.jpg "what is the difference between those two pictures"
The two pictures you provided appear to be identical. There are no visible differences between them. Both show the same character in the same pose with the same lighting and background.
```

You can also set the default model to GPT-4o or GPT-4 Vision by setting it in
the [configuration file](https://sgpt.readthedocs.io/en/stable/configuration/).

**Important:** The GPT-4o and GPT-4-vision API integration is currently in beta and may change in the future.
