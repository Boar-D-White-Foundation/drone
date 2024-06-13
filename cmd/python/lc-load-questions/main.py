import json
import time

from selenium import webdriver
from selenium.webdriver.chrome import service, options
from selenium.webdriver.common.by import By


_QUESTIONS_URL = "https://leetcode.com/problemset/algorithms"


def main():
    opts = options.Options()
    driver = webdriver.Chrome(
        options=opts,
        service=service.Service(
            executable_path="./chromedriver",
            port=8787,
        ),
    )
    driver.implicitly_wait(0.5)

    driver.get(_QUESTIONS_URL)
    time.sleep(5)
    questions_per_page_btn = driver.find_element(By.XPATH, "//*[starts-with(@id, 'headlessui-listbox-button')]")
    questions_per_page_btn.click()
    time.sleep(2)
    per_page_100_btn = driver.find_element(By.CSS_SELECTOR, "ul[role='listbox'] li:last-child")
    per_page_100_btn.click()
    time.sleep(2)

    data = []
    page = 0
    while True:
        rows = driver.find_elements(By.CSS_SELECTOR, "div[role='rowgroup'] div[role='row']")
        for i, row in enumerate(rows):
            if not page and not i:
                continue

            cols = row.find_elements(By.CSS_SELECTOR, "div[role='cell']")
            raw_name = cols[1].find_element(By.CSS_SELECTOR, "a").text.split(". ", 1)
            parsed_row = {
                "id": raw_name[0],
                "is_premium": bool(cols[0].find_elements(By.CSS_SELECTOR, "svg.text-brand-orange")),
                "link": cols[1].find_element(By.CSS_SELECTOR, "a").get_attribute("href"),
                "name": raw_name[1],
                "acceptance": float(cols[3].text.strip("%")),
                "difficulty": cols[4].text,
            }
            assert parsed_row["id"], parsed_row
            assert parsed_row["link"], parsed_row
            assert parsed_row["name"], parsed_row
            assert parsed_row["difficulty"] in ("Easy", "Medium", "Hard"), parsed_row
            data.append(parsed_row)

        next_btn = driver.find_element(By.CSS_SELECTOR, "nav[role='navigation'] button[aria-label='next']")
        if next_btn.get_attribute("disabled"):
            break

        next_btn.click()
        time.sleep(5)
        page += 1

    print(json.dumps(data, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
