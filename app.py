#!/usr/bin/env python
# -*- coding: utf-8 -*-

import sys
import posttriage
import pretriage


if __name__ == '__main__':
    if len(sys.argv) > 1 and sys.argv[1] == 'pretriage':
        pretriage.run()
    elif len(sys.argv) > 1 and sys.argv[1] == 'posttriage':
        posttriage.run()
    else:
        sys.exit("Available commands are 'pretriage' and 'posttriage'.")
            
