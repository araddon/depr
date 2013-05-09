Go dependency resolution
------------------------------

Dependency resolver, ensure all unique repo's are updated, with support for specific branches/tags etc.   

**Features**

* Different src control, and on-disk paths.   These versions could be forks of ones on github/launchpad from different users, but retain same path on disk so go imports are the same.
* "frozen" at a known version/hash
* use different branches from master/default

This is a command line tool that reads a yaml file that lists dependencies and ensures those exist.  


Usage (yaml dependency file) that is placed in a single location, we use these inside our bin/main's by creating *.depr.yaml*::
    
    # simple, ensure package exists and is updated
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

    # specific branch
    - src:  github.com/mattbaird/elastigo#develop

Usage::
    
    depr  # run looking for *.depr.yaml* in current folder

    depr --config  ./path/to/.depr.yaml

    depr --no-clean   # allow un-clean non-commited changes in depencies


some other go packages for dependency
-------------------------------------------
  
    * https://github.com/gopack/gpk
    * https://github.com/brianm/godeps
    * https://github.com/kr/godep
    * http://www.gonuts.io/ https://groups.google.com/forum/#!msg/golang-nuts/cyt-xteBjr8/zSIAcABKtKAJ
    * https://github.com/divoxx/goproj
