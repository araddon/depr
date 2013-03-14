Go dependency resolution
------------------------------

While coordinating with others on same code base, across large projects it became difficult to ensure that we had the same versions of packages due to

* These versions could be forks of ones on github/launchpad from different users, but retain same path on disk so go imports are the same.
* "frozen" at a known version/hash
* use different branches from master/default

This is a command line tool that reads a yaml file that lists dependencies and ensures those exist.


Usage (yaml dependency file) that is placed in a single location, we use these inside our cmd's::
    
    # simple, ensure package exists
    - src: github.com/suapapa/hilog

    # Use a repository that differs on disk vs github
    # for usage when you have forked a repository
    - as:  github.com/mattbaird/elastigo
      src: github.com/araddon/elastigo 

    # source location, seperate package and branch
    - as:  github.com/mattbaird/elastigo
      src: github.com/araddon/elastigo 
      branch: newsearch

    # specific version
    - as:  github.com/mattbaird/elastigo
      src: github.com/araddon/elastigo
      hash:  d364f0fbe86

    # specific version
    - src:  github.com/mattbaird/elastigo#d364f0fbe86
