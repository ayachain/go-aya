print("aapp.lua script is running.")

function CreateFile(name)

    fi = io.open(name, "a")

    if fi == nil then
        return "Create File Faild."
    end

    fi:write("DefaultContent")
    fi:close()
    return "Create File Success."

end


function FullWriteTest()

    file = io.open("writetext.txt", "w")

    if file == nil then
        return "FullWriteTest: Open File error"
    end

    -- 输入字符串
    file:write("test io.write\n");

    -- 输入数字
    file:write(2016)

    -- 输入分隔符
    file:write(" ")

    -- 继续输入数字
    file:write(7)
    file:write(" ")
    file:write(23)
    file:write("\n")

    -- 继续输入其他类型
    file:write(tostring(os.date()))
    file:write("\n")
    file:write(tostring(file))

    -- 关闭文件
    file:close()

    --读取
    local fileread = io.open("writetext.txt", "r")
    local content = fileread:read("*a");
    print("file content is : \n")
    print(content)

    fileread:close()

    return content

end


function FullReadTest()

    --"n"*：读取一个数字，这是唯一返回数字而不是字符串的读取格式。
    --"a"*：从当前位置读取余下的所有内容，如果在文件尾，则返回空串""。
    --"l"*：读取下一个行内容，如果在文件尾部则会返回nil

    local fileread = io.open("writetext.txt", "r")
    print("ReadLine: " .. fileread:read("l") )
    print("ReadNum: " .. fileread:read("n") )
    print("ReadAll: " .. fileread:read("a") )

    fileread:close()

end

FullWriteTest()
FullReadTest()
