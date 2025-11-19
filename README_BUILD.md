# Збірка проекту з іконкою

## Швидка збірка

Просто запустіть:

```bash
.\build.bat
```

Цей скрипт автоматично:
1. Встановить `rsrc` якщо потрібно
2. Згенерує `rsrc.syso` з іконкою та маніфестом
3. Зберіть `cid_retranslator.exe`

## Ручна збірка

### 1. Встановіть rsrc (один раз)

```bash
go install github.com/akavel/rsrc@latest
```

### 2. Згенерувати rsrc.syso

```bash
rsrc -ico icon.ico -manifest multiplepages.exe.manifest -o rsrc.syso
```

### 3. Зберіть програму

```bash
go build -o cid_retranslator.exe .
```

## Що знаходиться в ресурсах

- **icon.ico** - іконка програми (використовується для exe та системного трея)
- **multiplepages.exe.manifest** - Windows manifest з налаштуваннями DPI та візуальних стилів
- **resources.rc** - resource script (для довідки, rsrc використовує параметри командного рядка)
- **rsrc.syso** - згенерований файл з вбудованими ресурсами (автоматично включається в збірку)

## Примітки

- Файл `rsrc.syso` автоматично підхоплюється Go компілятором
- Іконка вбудовується прямо в `.exe` файл
- Після збірки `.exe` матиме правильну іконку в Провіднику та системному треї
- Маніфест забезпечує правильну підтримку DPI та візуальних стилів Windows

## Альтернатива: go-winres

Якщо `rsrc` не працює, можна використати `go-winres`:

```bash
# Встановити
go install github.com/tc-hib/go-winres@latest

# Створити JSON конфігурацію
go-winres init

# Відредагувати winres.json та згенерувати ресурси
go-winres make

# Зберіть
go build -o cid_retranslator.exe .
```
