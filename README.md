# account-managment-get-http-body-request

**Curl:** 

    curl --location --request GET 'localhost:8080/v1/users/123456' \
    --header 'Content-Type: application/json' \
    --data '{
    "DisplayName":"angelouz"
    }'