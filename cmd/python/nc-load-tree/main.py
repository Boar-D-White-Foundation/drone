from selenium import webdriver
from selenium.webdriver.chrome import service, options
from selenium.webdriver.common.by import By


def main():
    opts = options.Options()
    opts.add_argument("--headless=new")
    driver = webdriver.Chrome(
        options=opts,
        service=service.Service(
            executable_path="./chromedriver",
            port=8787,
        ),
    )
    driver.get("https://neetcode.io/roadmap")
    driver.implicitly_wait(0.5)
    text_box = driver.find_element(by=By.ID, value="1")
    text_box.click()
    trs = driver.find_elements(by=By.CSS_SELECTOR, value="tr")
    link = (
        trs[1]
        .find_elements(By.CSS_SELECTOR, "td")[2]
        .find_element(By.CSS_SELECTOR, "a")
        .get_attribute("href")
    )
    driver.quit()
    print(link)


if __name__ == "__main__":
    main()
