# "AI" Disclosure

This project is built with "AI"-augmented-engineering by leveraging LLMs to
improve development times. This document discloses how it is used, and how
this project aims to be transparent.

## Process

The design of this library is entirely **human-guided and authored**. Every
interface, interaction, and design composition is deliberately architected by
the author _first_. This design is prompted to a language model to produce an
implementation plan so that the design can be implemented.

Care is taken to verify:

* The interface contract is correct
* Tests validate the expectations of each interface, abstraction, etc.
* Documentation is correct, up-to-date, and not verbose, etc.

Things that do not pass this bar are sent back and iterated on. A human is
central in this development loop so that this project does not become, or
devolve into, pure slop.

Finishing touches and changes are typically authored by hand before committing
the changes.

## Transparency

In an effort to be transparent, as of [62d91c4f], this project attaches llm
provenance information as [git notes] to each commit. Notes are stored on
GitHub for full transparency. Notes are manual to observe and in order to be
inspected require the user to run:

```bash
git fetch origin refs/notes/*:refs/notes/*
```

Once fetched, notes are stored in two different namespaces:

* `llm`: this namespace indicates the meta-information about the prompt itself,
  such as `Model: claude-opus-4.8`

* `prompt`: this indicates the exact prompt provided to the LLM to start the
  session, along with any significant alteration instructions. Minor adjustments
  are curated and omitted.

These can be viewed in several ways:

1. Using `git log --show-notes=<namespace>`, e.g. `git log --show-notes=prompt`
   will pull up the prompt alongside the commits

2. Using `git notes --ref=<<namespace> show [commit]`, e.g. `git notes --ref=prompt`
   will fetch the current prompt, or `git notes --ref=prompt deadbeef` will get
   the note for the SHA `deadbeef`.

This project does not aim to misrepresent the AI-assisted nature of its origin,
and strives to be transparent about it. If you're a contributor considering
adding to this project, please follow the same pattern here.

[git notes]: https://git-scm.com/docs/git-notes
[62d91c4f]: https://github.com/bitwizeshift/go-cli/commit/62d91c4f0acdf91216b2a69ebc27b56fb1de81d5

## Commits

Git commit messages are essential to understanding the motivations to changes,
as well as any design-decisions that went into the planning behind it. As such,
**all commits are human-authored, no exception**. It is essential that real
rationale be provided for decisions to help better understand how a project
grows and changes over time. Language models, simply, **do not do this well**.
They summarize _what_ and sometimes _how_ it was done and don't provide any
real reason to _why_. They often over-describe what already exists, while
lacking critical information for the decision-making that lead to the change.

Simply put: that is not okay, and this project does not do this.
