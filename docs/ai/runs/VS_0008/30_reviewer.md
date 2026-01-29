# VS_0008 30_reviewer

## Summary

This is a good “bridge ticket” that makes VS_0006 optimizer actually usable:
- It produces `reward` labels from post-alert persistence windows (repo actionability proxy).
- It does not require network for acceptance thanks to mock mode.

## Issues

### Blocking

- None.

### Major

- `paper_log` timestamp is per-batch (`paper_log` uses one `now` for all events in a batch). This is OK for a proxy label, but later we may want per-event timestamps for tighter labeling.

### Minor

- Current labeler writes late-by seconds and uses grace; this is good. In a future ticket we can add a “missed window” label (instead of skipping) if you want full accounting.

