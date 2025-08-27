# Быстрый старт на Intel i5-6600 (x86_64 Linux)

## Минимальные требования
- Ubuntu 20.04+ или другой Linux x86_64
- gcc/g++ установлен
- Go 1.22+ установлен

## Быстрая установка Go (если не установлен)
```bash
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

## Сборка и запуск (3 команды)

```bash
# 1. Делаем скрипт сборки исполняемым
chmod +x build_x86.sh

# 2. Собираем майнер
./build_x86.sh

# 3. Запускаем
./start_x86.sh -u i5-6600 -t 4
```

## Что происходит?

1. **Подключение к пулу**: 146.103.50.122:5001
2. **Воркер**: 4BdyC3wW6BJiqCNp9Tdr2D9gVnBiVfFnCH.i5-6600
3. **Алгоритм**: Cuckoo Cycle 42
4. **Потоки**: 4 (оптимально для 4-ядерного i5-6600)

## Мониторинг

В отдельном терминале:
```bash
# CPU загрузка
htop

# Температура (требуется lm-sensors)
watch -n 1 sensors
```

## Проверка работы

Если всё работает правильно, вы увидите:
```
✓ Connected and authorized
✓ New work received
✓ Miner stats: cycles/s: XXXXX
```

## Проблемы?

### "libcuckoo_lean.a not found"
```bash
cd solver/tromp && make clean && make
```

### "connection refused"
Проверьте подключение к интернету и файервол

### Низкая производительность
```bash
sudo cpupower frequency-set -g performance
```
