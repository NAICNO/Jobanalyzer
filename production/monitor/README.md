This is a small program that runs as a cron job on a reliable host, contacts the remote sonalyze
instance using the primitive REST interface, and just checks whether it is up.  If it is not up, it
sends an email to an address.  It limits the amount of mail to one per hour.

