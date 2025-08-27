# Исправление ошибок сборки на x86

## Проблема решена!

Ошибки компиляции были вызваны двойным включением заголовков. Исправления:

1. ✅ Удалено дублирующееся включение `cuckoo.h`
2. ✅ Исправлен вызов `sipnode` для использования правильной функции из `lean.hpp`
3. ✅ Использована функция `setheader` вместо прямого вызова `blake2b`

## Теперь сборка должна работать:

```bash
# На x86_64 Linux:
tar -xzf go-cuckoo-miner-x86.tar.gz
cd go-cuckoo-miner
./build_x86.sh
```

## Если всё ещё есть ошибки с g++:

### Вариант 1: Использовать более общие флаги оптимизации
```bash
cd solver/tromp
make clean
export CFLAGS="-O3 -march=native -mavx2 -msse4.2"
export CXXFLAGS="-O3 -march=native -mavx2 -msse4.2 -std=c++14"
make
```

### Вариант 2: Отключить специфичные для Skylake оптимизации
Отредактируйте `build_x86.sh` и замените:
```bash
-march=skylake -mtune=skylake
```
на:
```bash
-march=native
```

### Вариант 3: Компилировать напрямую
```bash
cd solver/tromp
g++ -O3 -march=native -std=c++14 -pthread \
    -I. -Icuckoo-orig/src -Icuckoo-orig/src/crypto \
    -c cuckoo_lean.cpp -o cuckoo_lean.o
ar rcs libcuckoo_lean.a cuckoo_lean.o
```

## Структура файлов:

- `cuckoo_lean.cpp` - для x86_64 (полный алгоритм)
- `cuckoo_simple.cpp` - для ARM/Mac (упрощённая версия)
- Makefile автоматически выбирает правильный файл

## Проверка архитектуры:
```bash
uname -m  # должно показать x86_64
```

Архив `go-cuckoo-miner-x86.tar.gz` содержит все исправления!
