# goal · corpus.docs

- **State:** idle
- **Task:** none
- **Branch:** none

## Tasks

1. [ ] review
2. [ ] curate
3. [ ] snapshot
4. [ ] baseline

## Decisions

- plan: the corpus is a private GitHub repository with Git LFS, one git tag per version, and
  `MANIFEST.json` carries `version`; acceptance runs cite it.
- plan: curate first and push one v1, to spare LFS traffic and diffs.
- plan: a sample stays when it exercises a library scenario no other sample covers; gaps are
  filled by sourced or generated samples alike, judged on quality and fit, never hand-built; no
  size cap.
- plan: layout by format directory, with `covers` as a fixed-vocabulary scenario index.
- plan: `text/txt/ssh-session-log.txt` is redacted before any commit.
- plan: curation builds its set beside ~/docs; ~/docs stays untouched until v1.ai.inference's
  acceptance ships, and then becomes a clone of this repository at v1.

## Pending edits

- coordinator · `context/corpus-docs.md`: the ~/docs review note (what the corpus holds, how it
  changes, how it is shared), cited from `context/README.md`'s Notes.
