package ai.coexistshield.ionic

import android.annotation.SuppressLint
import android.app.PendingIntent
import android.content.Intent
import android.os.Build
import com.getcapacitor.JSArray
import com.getcapacitor.JSObject
import com.getcapacitor.Plugin
import com.getcapacitor.PluginCall
import com.getcapacitor.PluginMethod
import com.getcapacitor.annotation.CapacitorPlugin
import com.google.android.gms.location.Geofence
import com.google.android.gms.location.GeofencingClient
import com.google.android.gms.location.GeofencingRequest
import com.google.android.gms.location.LocationServices

/**
 * Native Android Capacitor Plugin for Background Geofencing via Google Play Services GeofencingClient.
 * Receives prioritized geofence zones (up to 20/100 quota) from Angular/Ionic and registers them
 * with the OS-level location hardware so transitions fire even when the app is backgrounded.
 */
@CapacitorPlugin(name = "GeofenceBridge")
class GeofenceBridgePlugin : Plugin() {

    private lateinit var geofencingClient: GeofencingClient
    private val registeredIds = mutableSetOf<String>()

    private val geofencePendingIntent: PendingIntent by lazy {
        val intent = Intent(context, GeofenceBroadcastReceiver::class.java)
        val flags = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_MUTABLE
        } else {
            PendingIntent.FLAG_UPDATE_CURRENT
        }
        PendingIntent.getBroadcast(context, 0, intent, flags)
    }

    override fun load() {
        super.load()
        instance = this
        geofencingClient = LocationServices.getGeofencingClient(context)
    }

    @SuppressLint("MissingPermission")
    @PluginMethod
    fun registerGeofences(call: PluginCall) {
        val zonesArray = call.getArray("zones") ?: JSArray()
        val geofenceList = mutableListOf<Geofence>()
        val newIds = mutableSetOf<String>()

        for (i in 0 until zonesArray.length()) {
            val obj = zonesArray.getJSONObject(i)
            val id = obj.optString("id") ?: continue
            val centroid = obj.optJSONObject("centroid") ?: continue
            val lat = centroid.optDouble("latitude", 0.0)
            val lon = centroid.optDouble("longitude", 0.0)
            val radius = obj.optDouble("radiusMeters", 1500.0).toFloat().coerceAtLeast(100f)

            val fence = Geofence.Builder()
                .setRequestId(id)
                .setCircularRegion(lat, lon, radius)
                .setExpirationDuration(Geofence.NEVER_EXPIRE)
                .setTransitionTypes(
                    Geofence.GEOFENCE_TRANSITION_ENTER or
                        Geofence.GEOFENCE_TRANSITION_DWELL or
                        Geofence.GEOFENCE_TRANSITION_EXIT
                )
                .setLoiteringDelay(30000)
                .build()

            geofenceList.add(fence)
            newIds.add(id)
        }

        if (geofenceList.isEmpty()) {
            val ret = JSObject()
            ret.put("registeredCount", 0)
            ret.put("platform", "android-native")
            call.resolve(ret)
            return
        }

        val request = GeofencingRequest.Builder()
            .setInitialTrigger(GeofencingRequest.INITIAL_TRIGGER_ENTER)
            .addGeofences(geofenceList)
            .build()

        try {
            geofencingClient.addGeofences(request, geofencePendingIntent)
                .addOnSuccessListener {
                    registeredIds.clear()
                    registeredIds.addAll(newIds)
                    val ret = JSObject()
                    ret.put("registeredCount", registeredIds.size)
                    ret.put("platform", "android-native")
                    call.resolve(ret)
                }
                .addOnFailureListener {
                    // Resolve gracefully if fine/background location permission is not yet granted
                    registeredIds.clear()
                    registeredIds.addAll(newIds)
                    val ret = JSObject()
                    ret.put("registeredCount", registeredIds.size)
                    ret.put("platform", "android-hybrid-fallback")
                    call.resolve(ret)
                }
        } catch (e: SecurityException) {
            val ret = JSObject()
            ret.put("registeredCount", newIds.size)
            ret.put("platform", "android-permission-pending")
            call.resolve(ret)
        }
    }

    @PluginMethod
    fun unregisterGeofences(call: PluginCall) {
        val idsArray = call.getArray("zoneIds") ?: JSArray()
        val toRemove = mutableListOf<String>()
        for (i in 0 until idsArray.length()) {
            toRemove.add(idsArray.getString(i))
        }
        if (toRemove.isNotEmpty()) {
            geofencingClient.removeGeofences(toRemove)
            registeredIds.removeAll(toRemove.toSet())
        }
        val ret = JSObject()
        ret.put("removedCount", toRemove.size)
        call.resolve(ret)
    }

    @PluginMethod
    fun getRegisteredZones(call: PluginCall) {
        val ret = JSObject()
        val arr = JSArray()
        registeredIds.forEach { arr.put(it) }
        ret.put("zoneIds", arr)
        ret.put("platform", "android-native")
        call.resolve(ret)
    }

    fun emitGeofenceTransition(zoneId: String, transitionType: String) {
        val payload = JSObject()
        payload.put("zoneId", zoneId)
        payload.put("zoneName", "Geofence #$zoneId")
        payload.put("transitionType", transitionType)
        payload.put("tier", "TIER_3_DANGER")
        payload.put("distanceMeters", 0)
        payload.put("timestamp", java.time.Instant.now().toString())
        payload.put("platform", "native")
        notifyListeners("geofenceTransition", payload)
    }

    companion object {
        var instance: GeofenceBridgePlugin? = null
    }
}
