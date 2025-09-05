### How to scan a repository

* Write an overview of the following `README.md`:

```markdown
<ATTACH_README_HERE>
```

* You must generate a **concise** summary for the `## Overview` subtitle of the
  resulting `AGENTS.md` file, and you must always include the primary title as a
  subtitle of the overview section.
* You should not include more than 3 paragraphs in the overview, excluding any
  potential build and/or usage instructions.
* If the `README.md` file contains build and/or usage instructions, make sure
  to include these in a subtitle of the overview, e.g. `### Build instructions`
  and/or `### Usage instructions`.
* You must search for any *commands* and *libraries* available in the repository
  and you may include these under the `### Library` and/or `### Commands`
  subtitles.
* You may first need to determine the primary programming language used in the
  repository by browsing through its source code files.
* You should not make up instructions, whatever instructions you provide must
  imperatively contain only relevant source code examples or commands and
  documentation from the repository itself.
* If the repository contains any usage *notes* and `Examples`, you may include
  these as well under the `### Examples` subtitle.

#### How to scan for commands and libraries

* In golang projects, commands are usually implemented in `cmd/`, whereas
  libraries are usually implemented in `api/` or `internal/`. Note that the
  resulting `AGENTS.md` file should only contain exported modules and functions.
* In C++ projects, commands are usually implemented in subfolders which contain
  a `main.cpp` source code file. Otherwise, the `README.md` file should usually
  mention some build instructions which permit to create the binaries. As for
  the libraries implemented in C++, they are usually implemented in subfolders
  which contain a `src/` or `devel/` subfolder of their own.

#### Example `AGENTS.md` file

```markdown
<ATTACH_EXAMPLE_AGENTS>
```