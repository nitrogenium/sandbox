# Инструкция по сборке для x86_64 Linux

## Системные требования

- Linux x86_64 (Ubuntu 20.04+ / Debian 11+ / CentOS 8+)
- Процессор: Intel i5-6600 или новее
- RAM: минимум 4GB (рекомендуется 8GB)
- Компиляторы: gcc/g++ 9.0+
- Go 1.22+

## 1. Подготовка системы

### Ubuntu/Debian:
```bash
# Обновляем систему
sudo apt update && sudo apt upgrade -y

# Устанавливаем необходимые пакеты
sudo apt install -y build-essential git wget gcc g++ make

# Устанавливаем Go 1.22
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### CentOS/RHEL/Fedora:
```bash
# Обновляем систему
sudo yum update -y

# Устанавливаем необходимые пакеты
sudo yum groupinstall -y "Development Tools"
sudo yum install -y git wget

# Устанавливаем Go 1.22
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

## 2. Клонирование и подготовка проекта

```bash
# Создаём рабочую директорию
mkdir -p ~/mining
cd ~/mining

# Клонируем проект (предполагаем, что у вас есть git репозиторий)
# Или копируем архив с проектом
cp -r /path/to/go-rebuild .
cd go-rebuild

# Проверяем версию Go
go version
# Должно показать: go version go1.22.0 linux/amd64
```

## 3. Сборка C++ солвера для x86_64

```bash
cd solver/tromp

# Удаляем старые файлы
make clean

# Для Intel i5-6600 (Skylake) используем оптимизации AVX2
export CFLAGS="-O3 -march=skylake -mtune=skylake -mavx2 -msse4.2"
export CXXFLAGS="-O3 -march=skylake -mtune=skylake -mavx2 -msse4.2 -std=c++14"

# Собираем библиотеку
make

# Проверяем, что библиотека создана
ls -la libcuckoo_lean.a
```

## 4. Сборка Go майнера

```bash
# Возвращаемся в корень проекта
cd ../..

# Скачиваем зависимости Go
go mod download

# Собираем майнер
go build -o bin/miner cmd/miner/main.go

# Делаем исполняемым
chmod +x bin/miner
```

## 5. Оптимизация для Intel i5-6600

### Настройка производительности:

```bash
# Отключаем энергосбережение
sudo cpupower frequency-set -g performance

# Проверяем частоты
cat /proc/cpuinfo | grep MHz

# Отключаем Turbo Boost для стабильности (опционально)
echo 1 | sudo tee /sys/devices/system/cpu/intel_pstate/no_turbo

# Настраиваем huge pages (2MB страницы)
echo 1024 | sudo tee /proc/sys/vm/nr_hugepages
```

### Создаём скрипт запуска:

```bash
cat > start_miner.sh << 'EOF'
#!/bin/bash

# Оптимизации для Intel i5-6600 (4 ядра)
export GOMAXPROCS=4
export MALLOC_ARENA_MAX=2

# Запускаем майнер
# -u : имя воркера (по умолчанию CPU-666)
# -t : количество потоков (по умолчанию 4 для i5-6600)

echo "Starting miner on Intel i5-6600..."
./bin/miner -u i5-6600 -t 4
EOF

chmod +x start_miner.sh
```

## 6. Запуск майнера

```bash
# Простой запуск
./bin/miner

# Запуск с кастомным именем воркера
./bin/miner -u MyWorker

# Запуск с 4 потоками (оптимально для i5-6600)
./bin/miner -t 4

# Запуск с отладкой
./bin/miner -debug

# Запуск через оптимизированный скрипт
./start_miner.sh
```

## 7. Мониторинг

### В отдельном терминале:
```bash
# Мониторинг CPU
htop

# Мониторинг температуры
watch -n 1 sensors

# Логи майнера (если перенаправлены)
tail -f miner.log
```

## 8. Systemd сервис (для автозапуска)

```bash
# Создаём сервис
sudo tee /etc/systemd/system/cuckoo-miner.service << EOF
[Unit]
Description=Cuckoo Cycle Miner
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME/mining/go-rebuild
ExecStart=$HOME/mining/go-rebuild/bin/miner -u i5-6600 -t 4
Restart=always
RestartSec=10

# Оптимизации
Environment="GOMAXPROCS=4"
Environment="MALLOC_ARENA_MAX=2"

# Лимиты
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

# Активируем и запускаем
sudo systemctl daemon-reload
sudo systemctl enable cuckoo-miner
sudo systemctl start cuckoo-miner

# Проверяем статус
sudo systemctl status cuckoo-miner

# Смотрим логи
sudo journalctl -u cuckoo-miner -f
```

## Устранение проблем

### 1. Ошибка сборки C++
```bash
# Проверьте версию gcc
gcc --version
# Должна быть 9.0 или выше

# Установите новый gcc если нужно
sudo apt install gcc-10 g++-10
sudo update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-10 100
```

### 2. Ошибка Go модулей
```bash
# Очистите кэш модулей
go clean -modcache
# Пересоздайте go.mod
go mod init github.com/nitrogen/go-miner
go mod tidy
```

### 3. Низкая производительность
```bash
# Проверьте governor
cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
# Должно быть "performance"

# Проверьте throttling
dmesg | grep -i thermal
```

## Ожидаемая производительность

Для Intel i5-6600 (4 ядра @ 3.3-3.9 GHz):
- Cycles/s: ~50-100k (зависит от оптимизации)
- Solutions/s: ~0.1-0.5 (зависит от сложности)
- Потребление: ~65W TDP

## Важные замечания

1. **Температура**: Следите за температурой CPU, она не должна превышать 80°C
2. **Память**: Майнер использует ~1-2GB RAM на поток
3. **Сеть**: Убедитесь, что порт 5001 не заблокирован файерволом
4. **Обновления**: Регулярно обновляйте майнер для лучшей производительности

## Контакты и поддержка

При возникновении проблем:
1. Проверьте логи: `./bin/miner -debug`
2. Убедитесь, что все зависимости установлены
3. Проверьте подключение к пулу: `telnet 146.103.50.122 5001`
