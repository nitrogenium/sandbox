# Сборка солвера для x86_64

## Важно!

`cuckoo_lean.cpp` - это версия для x86_64 с полным алгоритмом Tromp'а
`cuckoo_simple.cpp` - это упрощённая версия для ARM (Mac M1)

## Структура файлов:

- **Для x86_64 Linux**: используется `cuckoo_lean.cpp` с оптимизациями AVX2/SSE
- **Для ARM Mac**: используется `cuckoo_simple.cpp` (заглушка для тестирования)

## Как исправить ошибки компиляции на x86:

Файл `cuckoo_lean.cpp` уже исправлен для компиляции на x86_64:
1. Переименован конфликтующий `thread_ctx` → `solver_thread_ctx`
2. Исправлены типы переменных (int → uint32_t)
3. Исправлен вызов sipnode для 32-битной архитектуры
4. Удалена неиспользуемая переменная

## Компиляция на x86_64:

```bash
# На x86_64 Linux машине:
cd solver/tromp
make clean
make ARCH_FLAGS="-march=skylake -mtune=skylake -mavx2 -msse4.2" CPPFLAGS=""
```

## Если всё ещё есть ошибки:

Проверьте версию gcc:
```bash
gcc --version  # должна быть 9.0+
```

Убедитесь, что используется правильный файл:
```bash
# Makefile должен выбрать cuckoo_lean.cpp для x86_64
uname -m  # должно показать x86_64
```
