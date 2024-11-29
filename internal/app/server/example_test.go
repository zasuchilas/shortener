package server

func Example() {
	/*
		curl --request POST \
			--url http://localhost:8080/ \
			--header 'Content-Type: text/plain' \
			--header 'User-Agent: insomnia/9.3.3' \
			--cookie token=74fcbb99226ed7dcb3a468b7cdfd4dbd76ead5e4abba95232dd27c1455889cd52564 \
			--data humble-sauerkraut.name

		curl --request GET \
			--url http://localhost:8080/19xtf1u2 \
			--header 'User-Agent: insomnia/9.3.3' \
			--cookie token=74fcbb99226ed7dcb3a468b7cdfd4dbd76ead5e4abba95232dd27c1455889cd52564

		curl --request POST \
			--url http://localhost:8080/api/shorten \
			--header 'Content-Type: application/json' \
			--header 'User-Agent: insomnia/9.3.3' \
			--cookie token=74fcbb99226ed7dcb3a468b7cdfd4dbd76ead5e4abba95232dd27c1455889cd52564 \
			--data '{
					"url": "https://ya.ru"
				}'

		curl --request POST \
		   --url http://localhost:8080/api/shorten/batch \
		   --header 'Content-Type: application/json' \
		   --header 'User-Agent: insomnia/9.3.3' \
		   --cookie token=74fcbb99226ed7dcb3a468b7cdfd4dbd76ead5e4abba95232dd27c1455889cd52564 \
		   --data '[
		 	{
		 		"correlation_id": "batch1",
		 		"original_url": "https://ya.ru  "
		 	},
		 	{
		 		"correlation_id": "batch2",
		 		"original_url": "https://yandex.ru"
		 	},
		 	{
		 		"correlation_id": "batch3",
		 		"original_url": "http://ya.ru      "
		 	},
		 	{
		 		"correlation_id": "batch3(2)",
		 		"original_url": "http://ya.ru"
		 	},
		 	{
		 		"correlation_id": "batch5 (already used)",
		 		"original_url": "http://спорт.ru/"
		 	}
		 ]'

		curl --request GET \
		   --url http://localhost:8080/ping \
		   --header 'User-Agent: insomnia/9.3.3' \
		   --cookie token=74fcbb99226ed7dcb3a468b7cdfd4dbd76ead5e4abba95232dd27c1455889cd52564

		curl --request GET \
		   --url http://localhost:8080/api/user/urls \
		   --header 'User-Agent: insomnia/9.3.3' \
		   --cookie token=74fcbb99226ed7dcb3a468b7cdfd4dbd76ead5e4abba95232dd27c1455889cd52564

		curl --request DELETE \
		   --url http://localhost:8080/api/user/urls \
		   --header 'Content-Type: application/json' \
		   --header 'User-Agent: insomnia/9.3.3' \
		   --cookie token=74fcbb99226ed7dcb3a468b7cdfd4dbd76ead5e4abba95232dd27c1455889cd52564 \
		   --data '["19xtf1u5", "19xtf1u5", "19xtf1tt"]'
	*/
}
