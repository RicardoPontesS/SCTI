import http from 'k6/http';
import { sleep, check} from 'k6';
export const options = {
  vus: 100,
  duration: '120s',
};
export default function () {
        const url = ('http://localhost:8080/signup');
        const payload = JSON.stringify({Name: 'Teste', Email: 'teste@teste.com', Password:'123'})
        const headers = {
                'headers':{
                        'Content-Type':'aplication/json'
                }
        }
        const res = http.post(url,payload,headers)
        check (res, {'status should be 200':(r)=>r.status===200})
sleep(2);
}
