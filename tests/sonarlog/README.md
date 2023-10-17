# Catalog of test data files

Since JSON and CSV files can't easily hold comments about what they are for, this serves as a
catalog when necessary.  In most cases it's not necessary: just grep the .sh files for the name of a
data file to find the tests that use it.

* `whitebox-logclean.csv` is used by the whitebox tests in sonarlog/src/logclean.rs.
* `whitebox-config.json` is used by the the whitebox tests in sonarlog/src/configs.rs.
