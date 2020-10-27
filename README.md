# HamLog Converter for an Android HamLog App

This is a converter for [an Android HamLog App](https://play.google.com/store/apps/details?id=appinventor.ai_MzMd1494.HamLog). It converts the csv file which in the android storage root folder to the ADIF (.adi) format - output to `stdout`.

```bash
$ ./hamlog-c2a -h
Usage of ./hamlog-c2a:
  -c string
        station callsign (default "BG1REN")
  -i string
        input CSV filename (default "HamLog.csv")
  -s int
        start number
```

```bash
$ ./hamlog-c2a > log.adi
Converted 36 record(s) from 36 lines with 0 error(s). Exported 36 record(s).
Latest No: 36
```
