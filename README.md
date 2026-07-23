# Palkkatarkistus

Fyne-desktopapppi jolla vertaat TES-pohjaista palkkalaskentaa maksettuun palkkaan. Vuorot kalenteriin, TES tai omat arvot, laskelma aikavälille. Ei virallinen palkkaohjelma.

Go 1.23+, [Fyne](https://docs.fyne.io/started/).

```bash
make run
make run-demo   # esimerkkivuorot
make test
make build      # bin/payroll
make air        # live reload + demo-vuorot (vaatii airin)
```

## Paketointi (Linux / Mac / Windows)

Sovelluskuvake: `Icon.png`. Ohjeet: [PACKAGING.md](PACKAGING.md).

```bash
make package-linux    # tämä kone (Linux .tar.xz + pikakuvake)
make install-linux    # asenna ~/.localiin (näkyy sovellusvalikossa)
make uninstall-linux  # poista ~/.local-asennus
make cross-windows    # Windows (Docker + fyne-cross)
make cross-darwin     # macOS (Docker + SDK)
make cross-all        # kaikki alustat
```

MIT — `LICENSE`.
