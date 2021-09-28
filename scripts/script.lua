request = function()
    id_num = math.random(1,200)
    wrk.body="username=user" .. id_num .. "&password=password" .. id_num
    wrk.headers["Content-Type"] = "application/x-www-form-urlencoded"
    return wrk.format("POST", "/login")
end