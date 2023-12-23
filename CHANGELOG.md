# Changelog

- v0.1.7 - 2023-11-23
    - if the perameter bufRows in method Writer is 0, all data will be kept in memory
- v0.1.6 - 2023-11-23
    - set the upbound of the number of buffer rows as 1 million.
- v0.1.5 - 2023-11-24
    - replace tabs in cells with spaces.
- v0.1.4 - 2023-08-18
    - added a new style: StyleThreeLine (tree-line table).
- v0.1.3 - 2023-08-18
    - do not set hasHeader with true if all headers are empty strings.
    - added a new method: HasHeaders.
- v0.1.2 - 2023-06-27
    - fix setting a global MaxWidth short than cell texts.
- v0.1.1 - 2023-06-27
    - fix go.mod file
- v0.1.0 - 2023-06-27
    - first version
