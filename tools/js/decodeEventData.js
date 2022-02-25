function decodeDepositEventEventData(eventData) {
    // import {onchain_events} from '@starcoin/starcoin'
    // const {onchain_events} = require('@starcoin/starcoin')
    const eventName = 'DepositEvent';
    return onchain_events.decodeEventData(eventName, eventData).toJS();
}

const {onchain_events} = require('@starcoin/starcoin')
const eventName = 'DepositEvent';
const eventData = '0x00f2052a01000000000000000000000000000000000000000000000000000001035354430353544300';

//0x
//16
//1
//4096 256*16
//256
//1048576 256x256x16
//65536 256x256
//268465456 256*256*256*16
// 256*256*256
// 256*256*256*256*16
// 256*256*256*256*16
// 最大32位

console.log(onchain_events.decodeEventData(eventName,eventData).toJS());