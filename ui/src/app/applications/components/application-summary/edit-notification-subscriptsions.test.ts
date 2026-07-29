import {NOTIFICATION_SUBSCRIPTION_ANNOTATION_REGEX} from "./edit-notification-subscriptions";

test('rejects incorrect annotations', () => {
    expect(NOTIFICATION_SUBSCRIPTION_ANNOTATION_REGEX.test('apps_hanzo_ai/subscribe_a_b')).toEqual(false)
    expect(NOTIFICATION_SUBSCRIPTION_ANNOTATION_REGEX.test('apps.hanzo.ai/subscribe.a.b')).toEqual(true)
})
