# Source: https://github.com/dmachard/blocklist-domains

from pybadges import badge
from datetime import date

# create badge pour the date
b1 = badge(left_text='Last updated', 
          right_text="%s" % date.today(),
          right_color='blue')
          
with open("badge_date.svg", "w") as f:
    f.write(b1)
    
# create badge pour the number of hosts
with open("blocklist.txt", "r") as f:
    data = f.read()
domains = data.splitlines()

b2 = badge(left_text='Blocklisted hosts count',
           right_text="%s" % len(domains),
           right_color='red')
           
with open("badge_total.svg", "w") as f:
    f.write(b2)

# create badge wl the number of hosts
with open("allowlist.txt", "r") as f:
    data = f.read()
wldomains = data.splitlines()

a1 = badge(left_text='Whitelisted hosts count',
           right_text="%s" % len(wldomains),
           right_color='green')

with open("badge_total_allow.svg", "w") as f:
    f.write(a1)