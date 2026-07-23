# Packaging Palkkatarkistus (Linux / macOS / Windows)

## Quick start (this machine)

```bash
make fyne-tools          # installs fyne + fyne-cross if missing
make install-linux       # package + install to ~/.local (shows in app menu)
```

After install, search the app menu for **Palkkatarkistus**.

```bash
make uninstall-linux     # remove ~/.local install
make package-linux       # only build dist/Palkkatarkistus.tar.xz
```

Files installed:

- `~/.local/bin/Palkkatarkistus`
- `~/.local/share/applications/fi.palkkatarkistus.app.desktop`
- `~/.local/share/icons/...` + pixmaps icon

For day-to-day development keep using `make run` (uses embedded icon).

## Cross-platform builds (Docker)

From Linux (Docker required):

```bash
make cross-linux      # amd64 + arm64
make cross-windows    # .exe
make cross-darwin     # .app (needs macOS SDK for fyne-cross)
make cross-all
```

Outputs: `fyne-cross/dist/` (also copied to `dist/` when present).

### Notes

| Target  | Command              | Host requirement                          |
|---------|----------------------|-------------------------------------------|
| Linux   | `make package-linux` | this machine                              |
| Windows | `make cross-windows` | Docker                                    |
| macOS   | `make cross-darwin`  | Docker + Apple SDK, or build on a Mac     |

Metadata: `FyneApp.toml` · icon: `Icon.png` · ID: `fi.palkkatarkistus.app`  
Entry point: `./cmd/palkkatarkistus`

Bump `Version` / `Build` in `FyneApp.toml` before release packages.
