package stock

const (
	// 1、定制化脚本一（运营增加库存）
	IncrStockLua = `
    local goodStock
    local stockKey = KEYS[2]
    local incrStock = tonumber(ARGV[1])
    local val = redis.call('SET', KEYS[1], 1 , 'EX' , 3600 , 'NX')
    if  val == false  then
        return 0
    end
	if  val['ok'] == 'OK' then
        goodStock = redis.call('GET', stockKey)

        if goodStock == false then
		  return  redis.call('INCRBY', stockKey, incrStock)
		end

		if tonumber(goodStock) <= 0 then
          redis.call('SET', stockKey, incrStock)
		  return  incrStock
		end
        
     	return redis.call('INCRBY',stockKey, incrStock)
    else
     	return 0
    end
	`

	// 2、定制化脚本二（运营减少库存）
	DecrStockLua = `
    local val = redis.call('SET', KEYS[1], 1 , 'EX' , 3600 , 'NX')
    if val == false then 
        return 0
    end
	if  val['ok'] == 'OK' then
     	return redis.call('DECRBY', KEYS[2] , tonumber(ARGV[1]))
    else
     	return 0
    end
	`

	// 3、定制化脚本三（线上扣库存）
	ConsumerStockLua = `
    local  res = {}
    local  consumerStock = tonumber(ARGV[3])
    local  currStock = 0
    local  limitStock = 0
    for i = 1 , 2 , 1 do
        if  tonumber(ARGV[i]) > 0 then 
    		currStock = redis.call('INCRBY', KEYS[i] , consumerStock)
        	limitStock = tonumber(ARGV[i])
        	if (currStock < consumerStock) or (currStock - consumerStock > limitStock) then 
        		return res
        	end
    	end
    end

    local lowNum = currStock - consumerStock
	if lowNum < limitStock and currStock > 0 then
        local topNum = currStock
        if currStock > limitStock then 
            topNum = limitStock
        end
        for i = 1 , topNum - lowNum, 1 do
            res[i] =  ARGV[6]..':'..ARGV[7]..':'..ARGV[4]..':'..ARGV[5]..':'..i+lowNum
        end
        redis.call('SADD', KEYS[3] , unpack(res))
    end
    return res
    `

	// 4、定制化脚本四：从备份（防丢）缓存中获取库存ID，使用了redis.replicate_commands，请慎重使用该脚本
	ConsumerBackUpStockLua = ` 
    redis.replicate_commands()
    local  consumerStock = tonumber(ARGV[1])
    local  backUpStocks = redis.call('SRANDMEMBER', KEYS[1] , consumerStock)
	if backUpStocks and #backUpStocks > 0 then
        redis.call('SADD', KEYS[2] , unpack(backUpStocks))
        redis.call('SREM', KEYS[1] , unpack(backUpStocks))
    end
    return backUpStocks
    `

	// 5、定制化脚本五：
	UserStockIncrLua = `
    local  res = {}
    local  userLimitStock = tonumber(ARGV[1])
    local  consumerStock  = tonumber(ARGV[2])
    local  orderNum       = tonumber(ARGV[3])
    local  currStock = redis.call('INCRBY', KEYS[1] , consumerStock)

	local lowNum = currStock - consumerStock
    local index = 1
    local retryValue = ''
    if lowNum < userLimitStock and  currStock > 0 then 
    	local topNum = currStock
        if currStock > userLimitStock then 
            topNum = userLimitStock
        end
        for i = lowNum + 1 , topNum , 1 do
            if orderNum > 0 then 
				res[index] = ARGV[index+3]
            else
  				res[index] = ARGV[4]..':'..ARGV[5]..':'..ARGV[6]..':'..i
            end
			retryValue = retryValue..'$'..res[index]
            index = index + 1
        end
	end
	if #retryValue > 0 then
    	local val = redis.call('SET', KEYS[4], retryValue , 'EX' , 864000 , 'NX')
        if val == false or val['ok'] ~= 'OK' then 
			res = {}
			index = 1
		end
    end
    if orderNum >= index then 
		local  retrySet = {}
		for j = 1 , orderNum - index + 1, 1 do
			retrySet[j] = ARGV[index+2+j]
		end
		redis.call('SADD', KEYS[2] , unpack(retrySet))
        redis.call('SREM', KEYS[3] , unpack(retrySet))
	end
    return res
    `
)
