from webdriver_manager.chrome import ChromeDriverManager
from selenium.webdriver.chrome.service import Service

service = Service(ChromeDriverManager().install())
if "THIRD_PARTY_NOTICES."in service.path:
    service_path = service.path.replace("THIRD_PARTY_NOTICES.", "")
else:
    service_path = service.path
print(service_path, flush=True)
