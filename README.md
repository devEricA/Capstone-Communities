# Capstone-Communities
Senior Capstone project for Fall of 2021. The objective is to create an app that will allow people to join communities within their local area.

# Problem
Although the internet has promoted a new level of connectivity to others all around the globe, many find themselves still lost in the crowd. As services like Reddit and Discord foster online communities which grow into thousands of members, personal connections with others become harder to form as a result.

# Our Solution
The proposed product is an app that shows local communities within the user’s area. The user will be able to browse a map of his/her city, and will select the desired communities that he/she wishes to join. Once the user joins a community, news, posts, and events will pop up within the user’s feed that are related to the community. 

Our product solves the need by increasing the connectivity within the local scene. It will allow local communities to have a significantly easier time to connect with others of similar interests. By limiting the scope to the local space, we believe it will foster tighter knit communities.

# Instructions for Deployment
_These instructions are intended for Linux & Mac Users._

_Windows users should install [WSL](https://docs.microsoft.com/en-us/windows/wsl/install) and conduct the installation of this application within that system._
1. Install MariaDB: 
* [Linux Mint & Ubuntu](https://r00t4bl3.com/post/how-to-install-mariadb-10-3-on-linux-mint-19)
* [Other Systems](https://mariadb.com/downloads/)
2. Establish your Database server with user = root and password = root.
3. Run the SQL script(sqlCode.sql) in the root of this repo within your database server. 
4. If you haven't already, install Go [here](https://golang.org/doc/install)
5. If you haven't already, install Revel [by following this tutorial](https://revel.github.io/tutorial/gettingstarted.html)
6. Input <code>export PATH="$PATH:$GOPATH/bin"</code> into a unix terminal
7. Input <code> revel run Communities/</code>
8. Open your browser and navigate to localhost:9000.
