#import "touchid_darwin.h"
#import <LocalAuthentication/LocalAuthentication.h>
#import <Foundation/Foundation.h>

bool verify_touch_id() {
    @autoreleasepool {
        LAContext *context = [[LAContext alloc] init];
        NSError *error = nil;
        // Using LAPolicyDeviceOwnerAuthentication to allow fallback to system password if Touch ID is not available/opted-out
        if ([context canEvaluatePolicy:LAPolicyDeviceOwnerAuthentication error:&error]) {
            dispatch_semaphore_t sema = dispatch_semaphore_create(0);
            __block bool success = false;
            [context evaluatePolicy:LAPolicyDeviceOwnerAuthentication
                    localizedReason:@"Authenticate to access SSH credentials"
                              reply:^(BOOL s, NSError *e) {
                success = s;
                dispatch_semaphore_signal(sema);
            }];
            dispatch_semaphore_wait(sema, DISPATCH_TIME_FOREVER);
            return success;
        }
        return false;
    }
}
