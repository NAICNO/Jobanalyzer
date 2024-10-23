This is a small program that runs sonar repeatedly with sensible options for a given scenario.

There is some documentation about how to use this in [the sonalyze manual](../sonalyze/MANUAL.md),
look for the section "CRUDE, HIGH-LEVEL PROFILING".

Additionally, an email sent to a colleague:

De verktøyene vi har er bygget rundt en sampler (sonar, opprinnelig fra Radovan Bast ved UiT) og en
analysepakke (jobanalyzer, fra NAIC).  Normalt kjøres sonar via cron eller systemd på hver node ca
hvert 5 minutt og sender data til en vm hos oss, men den kan kjøres interaktivt eller via et
hjelper-program og lagre data lokalt, som du da kan analysere lokalt.

Sjekk ut https://github.com/NordicHPC/sonar.git
Sjekk ut https://github.com/NAICNO/Jobanalyzer.git

I sonar (du må ha en rimelig ny versjon av Rust installert, ihvertfall 1.65 tror jeg):

```
  $ cargo build --release
  $ target/release/sonar help
```

og deretter kan du sjekke at sonar virker som den skal og leverer data:

```
  $ target/release/sonar ps
```

I Jobanalyzer (du må ha Go 1.20 eller nyere, samt helst Rust):

```
  $ cd code
  $ make build
  $ sonalyze/sonalyze help
  $ sonard/sonard -h
```

Det du skal gjøre er å kjøre `sonard` i bakgrunnen mens du kjører applikasjonen din.  Sonard vil
kjøre sonar og vil lagre sample-data i en fil du spesifiserer.  Du kan deretter analysere de dataene
med `sonalyze`.

I utganspunktet (sample hvert sekund):

```
  $ sonard/sonard -i 1 -m 0 -s ../../sonar/target/release/sonar my-samples.csv &
```

og deretter kjører du programmet ditt

```
  $ python ...
```

og til slutt kan du avslutte samplingen:

```
  $ pkill sonard
```

og deretter kjøre analyse av loggen, fx

```
  $ sonalyze/sonalyze jobs -- my-samples.csv
```

Sonalyze har mange muligheter for spørring og formattering og datautvalk, du kan fx få en profil av
en jobb over tid (`sonalyze profile`).  Se code/sonalyze/MANUAL.md for endel informasjon.  Og spør,
for all del.  Ikke alt er intuitivt, på langt nær.
