# Collaborative Browser

The Collaborative Browser is a shared interface for many users. This project focuses on the design of intuitive _environments_ that offer proper affordances, not smart _users_ with extensive planning and reflection modules to accomplish complex tasks. Thus, if an incorrect action is chosen, it is not the user's fault. Rather, it is the environment's fault for not affording an appropriate action space.

## To run the shell

```bash
go run ./cmd/shell/shell.go -url scholar.google.com
```

This will start the shell with the starting location at [https://scholar.google.com/](https://scholar.google.com/). To interact with the browser or to see the changes visually, run:

```bash
go run ./cmd/shell/shell.go -url scholar.google.com -headful
```

Built-in commands:

- `help`: prints some usage instructions
- `headful`: if the current browser is running headless, open the headful representation
- `log`: logs the current browser and trajectory to the specified log path
- `exit`: gracefully exits the shell

## Markdown Browser

The Markdown Browser is an example of a text browser. It uses `virtual IDs` to enable textual users to select elements.
`Virtual IDs` provide a 1-to-1 mapping between input boxes, buttons, links, and other elements that may useful to select.
