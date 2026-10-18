#import <Foundation/Foundation.h>
#import <Capacitor/Capacitor.h>

CAP_PLUGIN(WearSyncPlugin, "WearSync",
    CAP_PLUGIN_METHOD(syncHudState, CAPPluginReturnPromise);
    CAP_PLUGIN_METHOD(sendImmediateAlert, CAPPluginReturnPromise);
)
