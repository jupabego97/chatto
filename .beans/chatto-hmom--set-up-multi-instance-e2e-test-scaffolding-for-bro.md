---
# chatto-hmom
title: Set up multi-instance E2E test scaffolding for Browse Spaces
status: todo
type: task
priority: high
created_at: 2026-03-24T17:10:41Z
updated_at: 2026-03-24T17:10:41Z
parent: chatto-e88o
---

The Browse Spaces E2E tests (browse-spaces.test.ts) only test single-instance scenarios. Need to:

- [ ] Add multi-instance test scaffolding to browse-spaces.test.ts (reuse fixtures from multi-instance-browse.test.ts)
- [ ] Add tests that verify instance name labels appear correctly with multiple instances
- [ ] Add tests that verify instance name labels are hidden with a single instance
- [ ] Consider consolidating browse-spaces.test.ts and multi-instance-browse.test.ts or ensuring they complement each other
