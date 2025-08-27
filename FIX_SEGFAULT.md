# Исправление Segmentation Fault

## Проблема

Майнер падает с segfault сразу после получения работы от пула. Это происходит при вызове C++ солвера.

## Вероятная причина

Неправильная сборка солвера или повреждённая библиотека.

## Решение

### 1. Проверьте, какой файл был скомпилирован:

```bash
cd solver/tromp
ls -la *.o
# Должен быть cuckoo_lean.o
```

### 2. Принудительно пересоберите правильную версию:

```bash
cd solver/tromp
make clean
rm -f *.o *.a

# Скомпилируйте правильный файл напрямую
g++ -O3 -march=native -std=c++14 -pthread \
    -I. -Icuckoo-orig/src -Icuckoo-orig/src/crypto \
    -c cuckoo_lean.cpp -o cuckoo_lean.o

# Создайте библиотеку
ar rcs libcuckoo_lean.a cuckoo_lean.o

# Проверьте
ar -t libcuckoo_lean.a
# Должно показать: cuckoo_lean.o
```

### 3. Пересоберите майнер:

```bash
cd ../..
go clean -cache
go build -o bin/miner cmd/miner/main.go
```

### 4. Тест с минимальными параметрами:

```bash
# Запустите с 1 потоком для отладки
./bin/miner -u test -t 1
```

## Альтернативное решение

Если проблема сохраняется, отредактируйте Makefile:

```bash
cd solver/tromp
nano Makefile
```

Найдите блок выбора источников и закомментируйте условие:
```makefile
# Source files
SOURCES = cuckoo_lean.cpp
```

## Отладка

Для более подробной диагностики:

```bash
# Включите core dumps
ulimit -c unlimited

# Запустите майнер
./bin/miner -u test -t 1

# Если создался core dump
gdb ./bin/miner core
(gdb) bt
(gdb) frame 0
(gdb) info locals
```

## Проверка памяти

Убедитесь, что достаточно памяти:
```bash
free -h
# Нужно минимум 4GB свободной RAM
```

## Важно!

`cuckoo_lean.cpp` - единственный солвер с полноценной реализацией алгоритма Cuckoo Cycle.
