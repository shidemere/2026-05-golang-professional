
Есть два варианта запустить тесты: 
1. Запустить тесты без каких либо ограничений чтобы в принципе увидеть сколько времени занял прогон, и какие ресурсы он за это время съел. Прогон выполняется за счёт `go test stats.go stats_optimization_test.go  -v`
2. Запустить тесты с ограничениями, чтобы убедиться, что задание пройдено. В таком случае они будут запускаться с каким то таймаутом и в единичном прогоне. Запустить тесты можно через `go test -v -count=1 -timeout=30s -tags bench .`

Обычный прогон даёт такие результаты
```shell
go test stats.go stats_optimization_test.go  -v
=== RUN   TestGetDomainStat_Time_And_Memory
    stats_optimization_test.go:46: time used: 672.822731ms / 300ms
    stats_optimization_test.go:47: memory used: 244Mb / 30Mb
    stats_optimization_test.go:49: 
                Error Trace:    /home/prth/Work/golang/2026-05-golang-professional/hw10_program_optimization/stats_optim
ization_test.go:49
                Error:          "672822731" is not less than "300000000"
                Test:           TestGetDomainStat_Time_And_Memory
                Messages:       the program is too slow
--- FAIL: TestGetDomainStat_Time_And_Memory (28.99s)
FAIL
FAIL    command-line-arguments  29.005s
FAIL

```

То есть тут результаты - тесты выполняются в два раза дольше положенного (672 мс из 300 максимальных) и по памяти занимает в 7 раз больше чем максимально может (244 мб против 30мб). 


### Заменить regexp.Match 

regexp.Match компилиируется при каждом вызове и так как это строка у нас вызывается в цикле мы при каждой новой итерации заново компилиируем регулярное выражение. 

Было
```go
func countDomains(u users, domain string) (DomainStat, error) {
	result := make(DomainStat)

	for _, user := range u {
		matched, err := regexp.Match("\\."+domain, []byte(user.Email))
		if err != nil {
			return nil, err
		}

		if matched {
			num := result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])]
			num++
			result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])] = num
		}
	}
	return result, nil
}
```

Стало
```go
func countDomains(u users, domain string) (DomainStat, error) {
	result := make(DomainStat)

	re := regexp.MustCompile("\\." + domain)
	for _, user := range u {
		matched := re.MatchString(user.Email)

		if matched {
			num := result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])]
			num++
			result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])] = num
		}
	}
	return result, nil
}
```

Результат
```shell
go test stats.go stats_optimization_
test.go  -v
=== RUN   TestGetDomainStat_Time_And_Memory
    stats_optimization_test.go:46: time used: 423.644812ms / 300ms
    stats_optimization_test.go:47: memory used: 112Mb / 30Mb
    stats_optimization_test.go:49: 
                Error Trace:    /home/prth/Work/golang/2026-05-golang-professional/hw10_program_optimization/stats_optim
ization_test.go:49
                Error:          "423644812" is not less than "300000000"
                Test:           TestGetDomainStat_Time_And_Memory
                Messages:       the program is too slow
--- FAIL: TestGetDomainStat_Time_And_Memory (11.19s)
FAIL
FAIL    com
```

Memory: 244 -> 112
Speed: 672 -> 423

Для интереса так же попробовал `strings.Contains()` но не получил никакого прироста, даже была деградация по скорости. 
Бенчмарк с использованием `strings.Contains()`:
```shell
go test stats.go stats_optimization_
test.go  -v
=== RUN   TestGetDomainStat_Time_And_Memory
    stats_optimization_test.go:46: time used: 465.556492ms / 300ms
    stats_optimization_test.go:47: memory used: 112Mb / 30Mb
    stats_optimization_test.go:49: 
                Error Trace:    /home/prth/Work/golang/2026-05-golang-professional/hw10_program_optimization/stats_optim
ization_test.go:49
                Error:          "465556492" is not less than "300000000"
                Test:           TestGetDomainStat_Time_And_Memory
                Messages:       the program is too slow
--- FAIL: TestGetDomainStat_Time_And_Memory (10.19s)
FAIL
FAIL    command-line-arguments  10.196s
FAIL
```

465 у contains VS 423 у MatchString
Оставляю MatchString.

### Замена io.ReadAll()

ReadAll считает все в память приложения, из за чего у нас может резко возрасти потребление по памяти. 
Посколько данные у нас построчные можно попробовать заменить на bufio.Scanner

Было 
```go
func getUsers(r io.Reader) (result users, err error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		var user User
		if err = json.Unmarshal([]byte(line), &user); err != nil {
			return
		}
		result[i] = user
	}
	return
}
```

Стало
```go
func getUsers(r io.Reader) (result users, err error) {
	scanner := bufio.NewScanner(r)
	i := 0
	for scanner.Scan() {
		line := scanner.Text()
		var user User
		if err = json.Unmarshal([]byte(line), &user); err != nil {
			return
		}
		result[i] = user
		i++
	}
	if err = scanner.Err(); err != nil {
		return
	}
	return
}
```

Потребление памяти снизилось 
```shell
go test stats.go stats_optimization_
test.go  -v
=== RUN   TestGetDomainStat_Time_And_Memory
    stats_optimization_test.go:46: time used: 411.198097ms / 300ms
    stats_optimization_test.go:47: memory used: 76Mb / 30Mb
    stats_optimization_test.go:49: 
                Error Trace:    /home/prth/Work/golang/2026-05-golang-professional/hw10_program_optimization/stats_optim
ization_test.go:49
                Error:          "411198097" is not less than "300000000"
                Test:           TestGetDomainStat_Time_And_Memory
                Messages:       the program is too slow
--- FAIL: TestGetDomainStat_Time_And_Memory (7.78s)
FAIL
FAIL    command-line-arguments  7.782s
FAIL
```

Memory: 112 -> 76

После этого я узнал что можно избавиться от преобразования в текст в виде `Scanner.Text()` и получать из сканнера данные напрямую в байтах. 
Функция стала выглядеть вот так
```go
func getUsers(r io.Reader) (result users, err error) {
	scanner := bufio.NewScanner(r)
	i := 0
	for scanner.Scan() {
		var user User
		if err = json.Unmarshal(scanner.Bytes(), &user); err != nil {
			return
		}
		result[i] = user
		i++
	}
	if err = scanner.Err(); err != nil {
		return
	}
	return
}
```

Потребление по памяти и скорости(!) тоже улучшилось
Результат 
```shell
go test stats.go stats_optimization_
test.go  -v
=== RUN   TestGetDomainStat_Time_And_Memory
    stats_optimization_test.go:46: time used: 360.701544ms / 300ms
    stats_optimization_test.go:47: memory used: 41Mb / 30Mb
    stats_optimization_test.go:49: 
                Error Trace:    /home/prth/Work/golang/2026-05-golang-professional/hw10_program_optimization/stats_optim
ization_test.go:49
                Error:          "360701544" is not less than "300000000"
                Test:           TestGetDomainStat_Time_And_Memory
                Messages:       the program is too slow
--- FAIL: TestGetDomainStat_Time_And_Memory (7.80s)
FAIL
FAIL    command-line-arguments  7.807s
FAIL

2026-05-golang-professional/hw10_program_optimization hw10_program_optimization  ✗ 
```

Memory: 76 -> 41
Speed: 423 -> 360

После этого я так же узнал что можно сделать json.Decoder котоырй берет данные прямо из файла, но особого прироста я от этого не получил. 

### Замена дефолтного json маршаллинга

Замена на easyJson наконец то позволила пройти порог по скорости. 
Сам код не особо сильно отличается (кроме того что метод анмаршаллинга применяется на структуре) поэтому код самой функции приводить не буду.
Необходимо было установить зависимость для работы с `easyJson` и сгенерировать дополнительный файлик с функциями. 
В общем кодогенерация оказалась быстрее рефлексии.

Результат
```go
go test stats.go stats_optimization_test.go stats_easyjson.go -v
=== RUN   TestGetDomainStat_Time_And_Memory
    stats_optimization_test.go:46: time used: 234.143402ms / 300ms
    stats_optimization_test.go:47: memory used: 31Mb / 30Mb
    stats_optimization_test.go:50: 
                Error Trace:    /home/prth/Work/golang/2026-05-golang-professional/hw10_program_optimization/stats_optim
ization_test.go:50
                Error:          "32688368" is not less than "31457280"
                Test:           TestGetDomainStat_Time_And_Memory
                Messages:       the program is too greedy
--- FAIL: TestGetDomainStat_Time_And_Memory (3.19s)
FAIL
FAIL    command-line-arguments  3.200s
FAIL
```

Memory: 76 -> 31
Speed: 423 -> 234

Осталось сэкономить еще один несчастный мегабайт

### Объединение countDomains и getUsers
Самое большое потребление по памяти, как я понимаю, это то, что мы создаем слайс размером в 100К. При этом нам почти наверняка никогда не понадобится такой размер. Можно было бы уменьшить слайс например. Но по сути этот слайс нужен только для того, чтобы передать из одной функции в другую данные. Поэтому можно обойтись вообще него если делать весь подсчет в рамках одной функции. 

Нам в принципе не нужен User кроме как для того, чтобы получить его Email. 
То есть после того как сделали Unmarshal можно взять емейл, положить в статистику и забыть юзера. 
Поэтому если объединить все в одной функции с вот таким кодом

```go
func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	result := make(DomainStat)
	scanner := bufio.NewScanner(r)
	re := regexp.MustCompile("\\." + domain)
	for scanner.Scan() {
		var user User

		if err := user.UnmarshalJSON(scanner.Bytes()); err != nil {
			return nil, err
		}

		matched := re.MatchString(user.Email)

		if matched {
			num := result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])]
			num++
			result[strings.ToLower(strings.SplitN(user.Email, "@", 2)[1])] = num
		}
	}

	return result, scanner.Err()
}
```

То получается сильный выигрыш по памяти. 
```shell
go test stats.go stats_optimizatio
n_test.go stats_easyjson.go -v
=== RUN   TestGetDomainStat_Time_And_Memory
    stats_optimization_test.go:46: time used: 292.396462ms / 300ms
    stats_optimization_test.go:47: memory used: 11Mb / 30Mb
--- PASS: TestGetDomainStat_Time_And_Memory (5.81s)
PASS
ok      command-line-arguments  5.812s

2026-05-golang-professional/hw10_program_optimization hw10_program_optimization  ? ❯ 
```

Memory: 31 -> 11
Speed: 234 -> 292

К сожалению чутка теряем в скорости, но все еще в пределах допустимого. 

### Добавил размер буффера
```go
scanner.Buffer(make([]byte, 64*1024), 1024*1024)
```

Memory: 11 -> 10
Speed: 292 -> 283 

### Финал 
Вот тут уже кажется что можно бы сдавать работу, но если запустить именно проверяющие тесты - они все равно показывают что финальное решение медленное. 

```go
go test -v -count=1 -timeout=30s -
tags bench .
=== RUN   TestGetDomainStat_Time_And_Memory
    stats_optimization_test.go:46: time used: 370.937439ms / 300ms
    stats_optimization_test.go:47: memory used: 11Mb / 30Mb
    stats_optimization_test.go:49: 
                Error Trace:    /home/prth/Work/golang/2026-05-golang-professional/hw10_program_optimization/stats_optim
ization_test.go:49
                Error:          "370937439" is not less than "300000000"
                Test:           TestGetDomainStat_Time_And_Memory
                Messages:       the program is too slow
--- FAIL: TestGetDomainStat_Time_And_Memory (6.00s)
FAIL
FAIL    github.com/shidemere/2026-05-golang-professional/hw10_program_optimization      6.002s
FAIL
```

Видимо что то кэшировалось до этого. 
Поэтому последнее что я могу придумать - не парсить целиком всего пользователя, а только его емейл. 

Поэтому я создал отдельную структуру с емейлом 

```go
type UserEmail struct { 
	Email string
}
```

и сгенерировал для неё код.

Запустил тесты и теперь все проходит
```shell
go test -v -count=1 -timeout=30s -tags bench .
=== RUN   TestGetDomainStat_Time_And_Memory
    stats_optimization_test.go:46: time used: 196.651598ms / 300ms
    stats_optimization_test.go:47: memory used: 3Mb / 30Mb
--- PASS: TestGetDomainStat_Time_And_Memory (3.02s)
PASS
ok      github.com/shidemere/2026-05-golang-professional/hw10_program_optimization      3.026s
```

