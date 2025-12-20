const BASE_URL = "https://nonadministrant-continuately-tamekia.ngrok-free.dev";

function onOpen() {
  const ui = SpreadsheetApp.getUi();
  ui.createMenu('Трекинг')
    .addItem('1. Регистрация', 'registerUser')
    .addItem('2. Войти', 'loginUser')
    .addItem('3. Отправить строку', 'sendTransaction')
    .addSeparator()
    .addItem('4. Получить отчет', 'getReport')
    .addItem('5. Установить бюджет', 'setBudget')
    .addItem('6. Получить бюджет', 'getBudgets')
    .addToUi();
}

function registerUser() {
  const ui = SpreadsheetApp.getUi();
  const email = ui.prompt('Регистрация', 'Введите Email:', ui.ButtonSet.OK).getResponseText();
  const pass = ui.prompt('Регистрация', 'Введите Пароль:', ui.ButtonSet.OK).getResponseText();

  if (!email || !pass) return;

  const payload = {
    email: email,
    password: pass
  };

  try {
    const options = {
      'method': 'post',
      'contentType': 'application/json',
      'payload': JSON.stringify(payload),
      'muteHttpExceptions': true
    };
    const response = UrlFetchApp.fetch(BASE_URL + "/register", options);
    ui.alert("Ответ сервера: " + response.getContentText());
  } catch (e) {
    ui.alert("Ошибка соединения: " + e.message);
  }
}

function loginUser() {
  const ui = SpreadsheetApp.getUi();
  const email = ui.prompt('Вход', 'Введите Email:', ui.ButtonSet.OK).getResponseText();
  const pass = ui.prompt('Вход', 'Введите Пароль:', ui.ButtonSet.OK).getResponseText();

  if (!email || !pass) return;

  const payload = {
    email: email,
    password: pass
  };

  try {
    const options = {
      'method': 'post',
      'contentType': 'application/json',
      'payload': JSON.stringify(payload),
      'muteHttpExceptions': true
    };
    const response = UrlFetchApp.fetch(BASE_URL + "/login", options);

    try {
      const json = JSON.parse(response.getContentText());

      if (json.token) {
        PropertiesService.getScriptProperties().setProperty('JWT_TOKEN', json.token);
        ui.alert("Успешный вход!");
      } else {
        ui.alert("Ошибка входа: " + response.getContentText());
      }
    } catch (parseErr) {
      ui.alert("Ошибка чтения ответа: " + response.getContentText());
    }

  } catch (e) {
    ui.alert("Ошибка соединения: " + e.message);
  }
}

function sendTransaction() {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getActiveSheet();
  const ui = SpreadsheetApp.getUi();
  
  const row = sheet.getActiveCell().getRow();
  
  const token = PropertiesService.getScriptProperties().getProperty('JWT_TOKEN');
  if (!token) { ui.alert("Сначала войдите в систему!"); return; }

  const amount = sheet.getRange(row, 1).getValue();
  const category = sheet.getRange(row, 2).getValue();
  const desc = sheet.getRange(row, 3).getValue();

  if (amount === "" || category === "") {
    ui.alert("Заполните Сумму и Категорию!");
    return;
  }

  const payload = {
    amount: parseFloat(amount),
    category: category.toString(),
    description: desc.toString()
  };

  const options = {
    'method': 'post',
    'contentType': 'application/json',
    'headers': {
      'Authorization': token,
      'ngrok-skip-browser-warning': 'true'
    },
    'payload': JSON.stringify(payload),
    'muteHttpExceptions': true
  };

  try {
    const response = UrlFetchApp.fetch(BASE_URL + "/transaction", options);
    const textResponse = response.getContentText();
    
    let json;
    try {
      json = JSON.parse(textResponse);
    } catch (e) {
      sheet.getRange(row, 4).setValue("Ошибка сети");
      return;
    }

    const cellStatus = sheet.getRange(row, 4);

    if (json.success === true) {
       cellStatus.setValue("Сохранено");
       cellStatus.setFontColor("green");
    } else {
       cellStatus.setValue(json.message); 
       cellStatus.setFontColor("red");
       cellStatus.setWrap(true);
       
       ui.alert("ОТКАЗ: " + json.message);
    }

  } catch (e) {
    sheet.getRange(row, 4).setValue("Err: " + e.message);
  }
}

function getReport() {
  const ui = SpreadsheetApp.getUi();
  const token = PropertiesService.getScriptProperties().getProperty('JWT_TOKEN');

  if (!token) {
    ui.alert("Сначала войдите в систему!");
    return;
  }

  const options = {
    'method': 'get',
    'headers': {
      'Authorization': token,
      'ngrok-skip-browser-warning': 'true' // <--- МАГИЧЕСКАЯ СТРОЧКА
    },
    'muteHttpExceptions': true
  };

  try {
    const response = UrlFetchApp.fetch(BASE_URL + "/report", options);

    if (response.getResponseCode() !== 200) {
      ui.alert("Ошибка: " + response.getContentText());
      return;
    }

    const json = JSON.parse(response.getContentText());

    let msg = "ВАШИ ТРАТЫ:\n\n";
    let total = 0;

    const cats = json.by_category || {};
    for (let key in cats) {
      msg += key + ": " + cats[key] + " руб.\n";
      total += cats[key];
    }

    msg += "\nИТОГО: " + total + " руб.";

    ui.alert(msg);

  } catch (e) {
    ui.alert("Ошибка соединения: " + e.message);
  }
}

function setBudget() {
  const ui = SpreadsheetApp.getUi();
  const token = PropertiesService.getScriptProperties().getProperty('JWT_TOKEN');
  if (!token) { ui.alert("Войдите!"); return; }

  const cat = ui.prompt('Бюджет', 'Категория (например, Еда):', ui.ButtonSet.OK).getResponseText();
  const limit = ui.prompt('Бюджет', 'Лимит (сумма):', ui.ButtonSet.OK).getResponseText();
  
  if (!cat || !limit) return;

  const payload = { category: cat, limit_amount: parseFloat(limit) };
  
  const options = {
    'method': 'post',
    'headers': { 'Authorization': token, 'ngrok-skip-browser-warning': 'true' },
    'contentType': 'application/json',
    'payload': JSON.stringify(payload)
  };

  UrlFetchApp.fetch(BASE_URL + "/set_budget", options);
  ui.alert("Бюджет установлен!");
}

function getBudgets() {
  const ui = SpreadsheetApp.getUi();
  const token = PropertiesService.getScriptProperties().getProperty('JWT_TOKEN');
  if (!token) return;

  const options = {
    'method': 'get',
    'headers': { 'Authorization': token, 'ngrok-skip-browser-warning': 'true' }
  };
  
  const resp = UrlFetchApp.fetch(BASE_URL + "/get_budgets", options);
  const json = JSON.parse(resp.getContentText());
  
  let msg = "БЮДЖЕТЫ:\n";
  if (json.budgets) {
    json.budgets.forEach(b => {
      msg += `${b.category}: ${b.limit_amount} р.\n`;
    });
  } else {
    msg += "Нет бюджетов";
  }
  ui.alert(msg);
}