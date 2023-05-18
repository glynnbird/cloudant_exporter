# Developing

In general, this should work as a normal Go application does.

## Version number

In local builds, the version number will usually show up as:

    <tag>-<commits since tag>-<latest commit SHA>

If your last commit was an annotated tag, it will just be that tag.

If there is no tag in the branch history, it's just the SHA.

## Publishing a new release

1. Create an annotated tag for the release:
   
       git tag -a v1.2

1. Push the tag up to GitHub. This will cause GH to build a
    docker image with the release.
