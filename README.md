# Collaborative Browser

The Collaborative Browser is a shared interface for many users. This project encourages development in AI to focus on building intuitive _environments_, not _users_, to do complex tasks.

## To run the shell

```bash
go run ./cmd/shell/shell.go -url scholar.google.com
```

This will start the shell with the starting location at [https://scholar.google.com/](https://scholar.google.com/). To interact with the browser or to see the changes, visually, run:

```bash
go run ./cmd/shell/shell.go -url scholar.google.com -headful
```

Use the `exit` command to gracefully exit the shell.

## Markdown Browser

The Markdown Browser is an example of a text browser. It uses `virtual IDs` to enable textual users to select elements.
`Virtual IDs` provide a 1-to-1 mapping between input boxes, buttons, links, and other elements that may useful to select.
