# tetra-mess

This is a simple CLI tool to take measurements of signal strength and quality from a TETRA
radio terminal through its peripheral equipment interface (PEI).

`tetra-mess` uses the proprietary AT command `AT+GCLI?` to query the TETRA terminal for
measurements. The command is only supported by radios from Motorola. You can find some more
infomation about the command in some forums and blogs (german):
- [TETRA, nützliche Einstellungen](https://berlinographics.com/funk/motorola-tetra-kein-bos/nutzliche-einstellungen/)
- [TETRA Netzabdeckung messen](https://www.funkmeldesystem.de/threads/61764-Tetra-Netzabdeckung-messen)
- [Tetra PEI: Software-Kommunikation AT-Komm.](https://hbaar.com/Homepage/?Funktechnik:Tetra_PEI:Software_-_Kommunikation_AT-Komm.)
- [Digitalfunk Tetra: Motorola PEI](https://hbaar.com/Homepage/?Funktechnik:Digitalfunk_%28Tetra%29:Motorola_PEI)

## Installation

You can install `tetra-mess` using the [Go](https://go.dev/) toolchain:

```bash
> go install github.com/ftl/tetra-mess@latest
```

## Usage

`tetra-mess` can record the measurements into a CSV or JSON file:

```bash
> tetra-mess measurements.csv
```

If you do not specify a filename, the measurements will be printed to the console.

`tetra-mess` can also evaluate a track file and convert it into a KML or GPX file in order
to visualize the measurements on a map:

```bash
> tetra-mess eval track measurements.csv --name "Today's test drive"
```

The default output format is KML, but you can use the GPX format with the flag `--format gpx`.

## License

This tool is published under the [GNU General Public License, Version 3](LICENSE)

Copyright [Florian Thienel](https://thecodingflow.com)
