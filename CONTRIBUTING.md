Want to contribute? Great! First, read this page.

### Before you contribute
Before we can use your code, you must sign the
[Google Individual Contributor License Agreement]
(https://cla.developers.google.com/about/google-individual)
(CLA), which you can do online. The CLA is necessary mainly because you own the
copyright to your changes, even after your contribution becomes part of our
codebase, so we need your permission to use and distribute your code. We also
need to be sure of various other things—for instance that you'll tell us if you
know that your code infringes on other people's patents. You don't have to sign
the CLA until after you've submitted your code for review and a member has
approved it, but you must do it before we can put your code into our codebase.
Before you start working on a larger contribution, you should get in touch with
us first through the issue tracker with your idea so that we can help out and
possibly guide you. Coordinating up front makes it much easier to avoid
frustration later on.

Contributions made by corporations are covered by a different agreement than
the one above, the
[Software Grant and Corporate Contributor License Agreement]
(https://cla.developers.google.com/about/google-corporate).

### Code reviews
All submissions, including submissions by project members, require review. We
use GitHub pull requests for this purpose.

### Code Style
We use Google style guides for all our code. Below are the guidelines for each
language we use.

#### Go
All Go source code must be formatted using gofmt and be lint-clean according to
golint. This will be checked by Travis but can be checked manually from a local
repo via `make gofmt golint`.

Code is expected to be gofmt- and lint clean before it is submitted for review.
Code reviews can then focus on structural details and higher level style
considerations. Many common mistakes are already documented in the
[Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
doc so it's worth being familiar with these patterns.

#### Python
All Python source code must be lint-clean according to pylint. This will be
checked by Travis but can be checked manually from a local repo via
`make pylint`.

Once code is pylint-clean, it can be submitted for review. In addition to lint
cleanliness, Python code must adhere to the
[Google Python Style Guide](https://google.github.io/styleguide/pyguide.html)
which has a number of additional conventions. Please be familiar with the style
guide and ensure code satisfies its rules before submitting for review.

##### Borrowed Standard Library Code
Standard library code that is borrowed from other open source projects such as
CPython and PyPy need not be lint clean or satisfy the style guide. The goal
should be to keep the copied sources as close to the originals as possible
while still being functional. For more details about borrowing this kind of
code from other places see the
[guidelines for integration](https://github.com/google/grumpy/wiki/Standard-libraries:-guidelines-for-integration).
