-- 测试3:类ERC20模式代币测试
local io = require("io")
local Json = require("json")

Interface = {

    name = {
        perfrom = name,

    }

}

_balanceTable = {}

function _loadBalanceTable()

    if not io.exist("/_balanceTable.table") then
        io.create("/_balanceTable.table")
    else
        _balanceTable = Json.decode(io.read("/_balanceTable.table"))
    end

end

function _saveBalanceTable()
    io.write( "/_balanceTable.table", Json.encode(_balanceTable), {t = true} )
end

-- ReadOnly
function name() return "AyaToken" end

-- ReadOnly
function totalSupply() return 100000000 end

-- ReadOnly
function decimals() return 1 end

-- ReadOnly
function symbol() return "AAY" end

-- ReadOnly
function balanceOf( req )
    _loadBalanceTable()
    return _balanceTable[req.address]
end

-- ReadAndWrite
function giveMeSomeToken( req )
    _loadBalanceTable()
    _balanceTable[req.address] = req.amount
    _saveBalanceTable()
end

-- ReadAndWrite
function transfer( req )
    _loadBalanceTable()
    _from = req.from
    _to = req.to
    _value = req.value

    if _balanceTable[_from] >= _value then
        _balanceTable[_from] = _balanceTable[_from] - _value

        if _balanceTable[_to] == nil then
            _balanceTable[_to] = _value
        else
            _balanceTable[_to] = _balanceTable[_to] + _value
        end
        _saveBalanceTable()
        return true
    else
        return false
    end
end

-- ReadOnly
function addressInfo( req )

    _loadBalanceTable()

    return {
        address = req.from,
        balance = _balanceTable[req.address]
    }

end
