#!/usr/bin/env python3
# Author: Yevgeniy Goncharov, https://lab.sys-adm.in
# Ref: https://github.com/dmachard/blocklist-domains
import os

from pybadges import badge
from datetime import date

# Work dir
BASE_DIR = os.path.dirname(os.path.abspath(__file__))

# Change work dir
os.chdir(BASE_DIR)
os.chdir("../public/files")


def create_date_badge():
    b1 = badge(left_text='Last updated',
               right_text="%s" % date.today(),
               right_color='blue')

    with open("badge_date.svg", "w") as f:
        f.write(b1)


def create_list_badge(list_name, list_description, file_badge, color_badge):
    # bl badge
    with open(list_name, "r") as f:
        list_data = f.read()
    list_elements = list_data.splitlines()

    badge_data = badge(left_text=list_description,
                       right_text="%s" % len(list_elements),
                       right_color=color_badge)

    with open(file_badge, "w") as f:
        f.write(badge_data)


# Generate badges
create_date_badge()
create_list_badge("bl.txt", "BL count", "badge_total_bl.svg", "red")
create_list_badge("bl_plain.txt", "BL with regex count", "badge_total_bl_plain.svg", "red")
create_list_badge("wl.txt", "WL count", "badge_total_wl.svg", "green")
create_list_badge("wl_plain.txt", "WL with regex count", "badge_total_wl_plain.svg", "green")
create_list_badge("ip_plain.txt", "IP count", "badge_total_ip.svg", "red")
