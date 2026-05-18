# scripts
Those are scripts to ease my job: deploying, testing and running.

Much of of it has been ~~shamelessly plundered~~ taken from [CausalMesh](https://github.com/eniac/causalmesh) scripts.

## deploying
The main idea is to follow the next steps:
1. deploy vms `[redis, gw, wkr1, wkr2, ..., wkrn]`
2. deploy redis
3. run experiment
    1. upload client code and restart redis
    2. run gw
    3. run wkrs
        1. run engine
        2. run launcher
    4. send requests
    5. retrieve results
4. 