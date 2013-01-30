Go dependency resolution
------------------------------

While coordinating with others on same code base, across large projects it became difficult to ensure that we had the same versions of packages due to

* These versions could be forks of ones on github/launchpad (but retain same path) 
* "frozen" at a known commit level/hash
* different branches from default

This is a command line tool that reads a yaml file that lists dependencies and ensures those exist.


TODO

* Allow src to be src you get from seperate from path
* Allow another branch from master
* allow a hash to be where you are frozen at
* In case of conflict/working branch not clean?

Usage (yaml dependency file) that is placed in a single location::
    
    # simple, ensure package exists
    - path: github.com/suapapa/hilog

    # Source location, seperate package location 
    - path:  github.com/mattbaird/elastigo
      src: github.com/araddon/elastigo 

    # source location, seperate package and branch
    - path:  github.com/mattbaird/elastigo
      src: github.com/araddon/elastigo 
      branch: newsearch

    # specific hash
    - path:  github.com/mattbaird/elastigo
      src: github.com/araddon/elastigo
      hash:  d364f0fbe86
      