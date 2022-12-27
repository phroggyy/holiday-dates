# Public Holidays Finder

This application lists the public holidays for all countries available through
the [Nager.Date API](https://date.nager.at/Api), between two years.

This can be useful in the financial industry (ensuring payout dates do not fall on a holiday),
on-call pay calculation (possibly applying higher rates for public holidays than weekdays), or
really any time you need to deal with public holidays for different countries.

It can be run either as a command-line application (e.g `./dates -start-year=2023 -years=2`
will get you all public holidays for 2023-2025), or as an HTTP server with
the `serve` subcommand: `./dates serve`. It will run on port `8080` (currently
not configurable), and accepts a path of `/:startYear/:endYear`, e.g `curl http://localhost:8080/2023/2025`.
