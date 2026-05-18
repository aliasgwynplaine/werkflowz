# -*- encoding: utf-8 -*-
#import fabric
from sys import argv

def main(gwaddr, wrks):
    # copy the folder
    # remote compile
    # remote execute
    #   gw
    #   for wkrs
    #     # engine
    #     # launchers
    print("gw address: ", gwaddr)
    print("worker list: ", wrks)
    pass

if __name__ == '__main__' :
    # check args
    if len(argv) < 3 :
        print(f"usage: %s <ip-gw> <wrks...>" % argv[0])
        exit(1)
    
    main(argv[1], argv[2:])
    pass