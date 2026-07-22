# Palkkatarkistus

Desktop-sovellus jolla voit tarkistaa, että palkkalaskelma vastaa tehtyjä tunteja ja Vartiointialan TES:n mukaista korvausta. Syötät vuorot kalenteriin, valitset TES-pohjan (tai omat arvot), ja lasket odotetun palkan aikavälille.

Käyttöliittymä on suomeksi. Laskenta on TES-pohjainen malli, ei virallinen palkkaohjelma — tarkoitus on antaa vertailuluku palkkalaskelmaa vasten.

## Mitä se tekee

- **Asetukset** — TES (Turvallisuusala / oma), taso, PK-seutu vs muu Suomi, palvelusaika, tuntipalkka ja lisät (ilta, yö, la, su, pyhä). Lisäksi koulutuslisä, muu lisä (tunneittain tai kiinteä), sekä jaksoylityön asetukset.
- **Vuorot** — kuukausikalenteri: lisää, muokkaa ja poista vuoroja. Yön yli menevät vuorot jakautuvat päiville. Päivän merkeissä näkyvät lisät (H/S/P/I/Y/L/50%/100%).
- **Laskelma** — valitse aikaväli, syötä poissaolotunnit jaksoa varten tarvittaessa, laske palkka. Näyttää erittelyn (pohja, lisät, pidennysylityö, jaksoylityö).
- **PDF-tuonti** ja **Vertailu** — paikat tulevalle työlle (ei vielä sisältöä).

### TES-logiikkaa jota malli tukee

- Tasopalkat 1.8.2025 (PKS / Muu Suomi), palvelusaikalisä, koulutuslisä 0,25 e/h
- Ilta / yö / lauantai / sunnuntai / pyhä
- **TES 31 §** — vuoro yli 12 h: ylimenevät tunnit kuten ylityö (50 %, ja jaksossa yli 18 h → 100 %). Nämä tunnit eivät kuulu jaksoylityöhön.
- **TES 29 §** — jaksoylityö: 120 h / 3 vk, tai tasoittuminen 128 / 112 h. Ensimmäiset 18 ylityötuntia 50 %, loput 100 %. Poissaolot (loma, sairaus, arkipyhä jne.) voi laskea mukaan tuntisummaan.

Asetukset elävät istunnossa; niitä ei vielä tallenneta levylle.

## Vaatimukset

- Go 1.23+
- Fyne-riippuvuudet (OpenGL / järjestelmän GUI-kirjastot). Linuxilla tyypillisesti tarvitaan mm. `libgl` ja X11/Wayland-kehityspaketit — katso [Fyne getting started](https://docs.fyne.io/started/).

## Käyttö

```bash
make run          # normaali käynnistys
make run-demo     # esimerkkivuorot (heinä–elo 2026) valmiina
make test
make build        # bin/payroll
```

Tai suoraan:

```bash
go run ./cmd/
go run ./cmd/ -demo
```

Live reload demolla (vaatii [air](https://github.com/air-verse/air)):

```bash
go install github.com/air-verse/air@latest
make air
```

## Rakenne

```
cmd/           käynnistys
internal/ui/   Fyne-käyttöliittymä, TES-profiilit, vuorokalenteri
internal/calc/ palkkalaskenta, pyhät, jakso- ja pidennysylityö
```

## License

MIT — katso `LICENSE`.
