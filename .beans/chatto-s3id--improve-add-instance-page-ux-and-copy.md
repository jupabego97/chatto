---
# chatto-s3id
title: Improve Add Instance page UX and copy
status: todo
type: task
priority: normal
created_at: 2026-03-25T08:31:09Z
updated_at: 2026-03-25T08:31:09Z
parent: chatto-wadw
---

The Add Instance page (/chat/instances/add) does two things — sign into the local instance and connect to a remote one — but neither is explained well. A new user has no idea what an 'instance' is or why they'd connect to another one.

## Problems

- No explanation of what Chatto is or how instances work
- 'Sign in to join the conversation' is vague — join what?
- 'or connect to another instance' assumes the user knows what an instance is
- 'Instance URL' is developer jargon
- The two sections look disconnected — no visual or conceptual bridge

## Suggestions

- Brief intro explaining Chatto is a federated chat platform where communities run their own servers
- The top section should clearly say this is YOUR server (the one hosting this app)
- The bottom section should explain that you can also connect to other Chatto servers, like joining a Discord server but self-hosted
- Use friendlier language: 'Server address' instead of 'Instance URL', 'Connect to a server' instead of 'connect to another instance'
- Consider showing the instance name more prominently (it's the admin-configured name, not just 'Chatto')
- The welcome message (if configured) should be visible here

## Context

This page is shown when the origin instance is detected but the user isn't authenticated. It's also the entry point for adding remote instances via the new OAuth flow.
