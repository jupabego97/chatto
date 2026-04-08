# Typing Indicators

## Overview

- When a user is composing a message, other users in the room see a typing indicator with the typer's avatar and animated dots.
- Typing indicators work for both room messages and thread replies (scoped independently).
- Typing indicators are live-only events — they are never persisted.

## Timing

- The client sends a typing event at most every 2 seconds (debounced).
- The receiving client displays the indicator for 6 seconds after the last typing event, then clears it.
- When a user actually posts a message, their typing indicator is immediately removed.
- The debounce resets after sending a message, so the next keystroke sends immediately.

## Scoping

- Typing indicators are scoped to a room and optionally a thread.
- The room view only shows indicators for users typing in the room (not in threads).
- A thread pane only shows indicators for users typing in that specific thread.

## Authorization

- Room membership is required to send and receive typing indicators. No additional permissions are needed.
