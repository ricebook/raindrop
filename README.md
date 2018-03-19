# raindrop

这个项目是受到了 Twitter [Snowflakes](https://github.com/twitter/snowflake) 项目的启发。但是目前 Snowflakes 已经停掉了，Twitter 准备把他做成给予 [Twitter-server](https://twitter.github.io/twitter-server/) 的一套服务。

这个雨滴（Raindrop）项目是使用 Golang 来实现的。关键思想是同 Snowflakes 一致。但是对一些应用场景进行了简化，更加适合一个项目早期使用。后面如果有需求，会逐渐服务化，支持多种方式的使用。

ps：这个代码写的有一段时间了，但是现在在赶别的项目，所以就不更新一些使用方法了。  
启动的话，直接执行 `./raindrop` 就行。  
